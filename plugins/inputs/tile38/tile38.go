package tile38

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/tidwall/gjson"
)

type Tile38 struct {
	Servers []string `toml:"servers"`
	Stats   bool     `toml:"keys_stats"`
}

var sampleConfig = `
## specify servers via a url matching:
## [:password]@address[:port]
## e.g.
##   localhost:9851
##   :password@192.168.100.100
##
## If no servers are specified, then localhost is used as the host.
## If no port is specified, 9851 is used
  servers = ["localhost:9851"]

## If true, collect stats for all keys. Default: false.
#  keys_stats = true
`

const defaultPort = "9851"

func (t *Tile38) SampleConfig() string {
	return sampleConfig
}

func (t *Tile38) Description() string {
	return "Read metrics from one or many Tile38 servers"
}

func (t *Tile38) Gather(acc telegraf.Accumulator) error {

	if len(t.Servers) == 0 {
		url := &url.URL{
			Scheme: "tcp",
			Host:   "localhost:9851",
		}
		t.gatherServer(url, acc)
		return nil
	}

	var wg sync.WaitGroup
	for _, serv := range t.Servers {
		if !strings.HasPrefix(serv, "tcp://") {
			serv = "tcp://" + serv
		}
		u, err := url.Parse(serv)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse to address '%s': %s", serv, err))
			continue
		}
		_, _, err = net.SplitHostPort(u.Host)
		if err != nil {
			u.Host = u.Host + ":" + defaultPort
		}
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			acc.AddError(t.gatherServer(u, acc))
		}(serv)
	}

	wg.Wait()
	return nil
}

func (t *Tile38) gatherServer(addr *url.URL, acc telegraf.Accumulator) error {
	tags := make(map[string]string)
	fileds := make(map[string]interface{})

	client := initTile38(addr)
	conn := client.Get()
	defer client.Close()

	info, err := redis.String(conn.Do("SERVER"))
	if err != nil {
		return err
	}

	host, port, _ := net.SplitHostPort(addr.Host)
	keysArray := [...]string{
		"id",
		"aof_size",
		"avg_item_size",
		"heap_released",
		"heap_size",
		"http_transport",
		"in_memory_size",
		"max_heap_size",
		"mem_alloc",
		"num_collections",
		"num_hooks",
		"num_objects",
		"num_points",
		"num_strings",
		"pid",
		"pointer_size",
		"read_only"}

	tags["server"], tags["port"] = host, port

	for _, k := range keysArray {
		val := gjson.Get(info, "stats."+k)
		switch val.Type {
		case gjson.String:
			tags[k] = val.String()
		case gjson.Number:
			fileds[k] = val.Int()
		case gjson.True:
			fileds[k] = int64(1)
		case gjson.False:
			fileds[k] = int64(0)
		default:
			acc.AddError(fmt.Errorf("Unable to parse json filed: '%s' value: '%s',type: '%s'", k, val, val.Type))
		}
	}

	acc.AddFields("tile38_server", fileds, tags)

	if t.Stats {
		k, err := redis.String(conn.Do("KEYS", "*"))
		if err != nil {
			return err
		}
		keys := gjson.Get(k, "keys").Array()

		for _, key := range keys {
			ktags := make(map[string]string)
			kfileds := make(map[string]interface{})
			jstr, err := redis.String(conn.Do("STATS", key))
			if err != nil {
				return err
			}
			stats := gjson.Get(jstr, "stats").Array()
			ktags["id"] = gjson.Get(info, "stats.id").String()
			ktags["server"], ktags["port"] = host, port
			ktags["key"] = key.String()

			kfileds["in_memory_size"] = gjson.Get(stats[0].String(), "in_memory_size").Int()
			kfileds["num_objects"] = gjson.Get(stats[0].String(), "num_objects").Int()
			kfileds["num_points"] = gjson.Get(stats[0].String(), "num_points").Int()
			kfileds["num_strings"] = gjson.Get(stats[0].String(), "num_strings").Int()
			acc.AddFields("tile38_stats", kfileds, ktags)
		}
	}
	return nil
}

func initTile38(addr *url.URL) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     1,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(addr.Scheme, addr.Host)
			if err != nil {
				return nil, err
			}
			if addr.User != nil {
				pwd, _ := addr.User.Password()
				if _, err := c.Do("AUTH", pwd); err != nil {
					c.Close()
					return nil, err
				}
			}
			if _, err := c.Do("OUTPUT", "JSON"); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func init() {
	inputs.Add("tile38", func() telegraf.Input { return &Tile38{} })
}
