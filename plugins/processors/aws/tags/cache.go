package tags

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/influxdata/telegraf"
)

// Tags is a read only map. Since it could be returned to multiple go routines simultaneously
// we have to make sure no one writes to it to avoid race conditions.
type Tags struct {
	t time.Time
	m map[string]string
}

func (t *Tags) Keys() (keys []string) {
	for key := range t.m {
		keys = append(keys, key)
	}
	return
}

func (t *Tags) Value(key string) (value string) {
	value = t.m[key]
	return
}

func (t *Tags) age() time.Duration {
	return time.Now().Sub(t.t)
}

type TagCache struct {
	maxAge         time.Duration
	requestTimeout time.Duration
	ec2Client      *ec2.Client
	log            telegraf.Logger
	// cache maps instance ids to a set of tags
	// TODO: one optimization that could be done, if memory is a concern, is to only save the tags that are requested
	cache     map[string]*Tags
	cacheLock sync.RWMutex
	// requests holds all ongoing requests so that calls can check for already ongoing requests
	// and simply wait on them instead of duplicating calls.
	requests     map[string]*sync.WaitGroup
	requestsLock sync.Mutex
	// requestsQueue is used to limit the amount of parallel requests
	// TODO: since we already limit parallelism through the use of the parallel package
	//       we could think about removing this. However, it would make more sense to increase
	//       the number of workers to a fixed, reasonable amount and only use this request limiting
	//       since the workers (most of the time) won't need to hit the network anyways.
	requestsQueue chan bool
}

// Get simplifies the access to tags by always choosing the quickest option to
// obtain the needed tags and automatically renewing outdated cache entries.
func (c *TagCache) Get(id string) *Tags {
	c.cacheLock.RLock()
	tags, ok := c.cache[id]
	c.cacheLock.RUnlock()
	if ok {
		c.log.Debugf("serving tags from cache for %s", id)
		if tags.age() > c.maxAge {
			c.log.Debugf("cache expired for %s, refreshing", id)
			go c.fetch(id)
		}
		return tags
	}

	c.fetch(id)

	c.cacheLock.RLock()
	tags = c.cache[id]
	c.cacheLock.RUnlock()
	return tags
}

// fetch either waits for the existing request to complete or triggers a new one and waits
// for it to complete.
// TODO: this can result in duplicate requests if someone is waiting at the top of the function while
//       we delete the WaitGroup at the bottom. However, that should generally not happen since we
//       wrote the value already to the cache which prevents anyone from calling fetch on this id.
func (c *TagCache) fetch(id string) {
	c.log.Debugf("fetching tags for %s", id)
	c.requestsLock.Lock()
	if req, ok := c.requests[id]; ok {
		c.requestsLock.Unlock()
		c.log.Debugf("request to fetch tags for %s already running, waiting", id)
		req.Wait()
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	c.requests[id] = wg
	c.requestsLock.Unlock()

	tags := c.getTags(id)
	c.cacheLock.Lock()
	c.cache[id] = &Tags{m: tags, t: time.Now()}
	c.cacheLock.Unlock()

	wg.Done()
	c.requestsLock.Lock()
	delete(c.requests, id)
	c.requestsLock.Unlock()
}

// getTags is the actual call to the AWS API to retrieve tags for a given
// instance id. Parallelism is limited by the requestsQueue
func (c *TagCache) getTags(id string) map[string]string {
	c.log.Debugf("calling aws API for %s, %d requests currently in flight", id, len(c.requestsQueue))
	// this will block if we already have maxParallel requests in flight
	c.requestsQueue <- true

	ctx := context.Background()
	context.WithTimeout(context.Background(), c.requestTimeout)
	describeTags, err := c.ec2Client.DescribeTags(ctx, &ec2.DescribeTagsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []string{id},
			},
		},
	})

	tags := make(map[string]string)
	if err != nil {
		c.log.Errorf("unable to describe tags for instance id %s; reason: %s", id, err.Error())
	} else {
		for _, tag := range describeTags.Tags {
			tags[*tag.Key] = *tag.Value
		}
	}

	<-c.requestsQueue
	return tags
}

func NewTagCache(maxAge, requestTimeout time.Duration, maxParallel int, ec2Client *ec2.Client, log telegraf.Logger) *TagCache {
	c := &TagCache{
		maxAge:         maxAge,
		requestTimeout: requestTimeout,
		ec2Client:      ec2Client,
		log:            log,
		cache:          make(map[string]*Tags),
		cacheLock:      sync.RWMutex{},
		requests:       make(map[string]*sync.WaitGroup),
		requestsLock:   sync.Mutex{},
		requestsQueue:  make(chan bool, maxParallel),
	}
	return c
}
