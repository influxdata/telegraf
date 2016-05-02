package mongodb

import (
	"log"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Server struct {
	Url        *url.URL
	Session    *mgo.Session
	lastResult *MongoStatus
}

func (s *Server) getDefaultTags() map[string]string {
	tags := make(map[string]string)
	tags["hostname"] = s.Url.Host
	return tags
}

func (s *Server) gatherData(acc telegraf.Accumulator) error {
	s.Session.SetMode(mgo.Eventual, true)
	s.Session.SetSocketTimeout(0)
	result_server := &ServerStatus{}
	err := s.Session.DB("admin").Run(bson.D{{"serverStatus", 1}, {"recordStats", 0}}, result_server)
	if err != nil {
		return err
	}
	result_repl := &ReplSetStatus{}
	err = s.Session.DB("admin").Run(bson.D{{"replSetGetStatus", 1}}, result_repl)
	if err != nil {
		log.Println("Not gathering replica set status, member not in replica set")
	}

	jumbo_chunks, _ := s.Session.DB("config").C("chunks").Find(bson.M{"jumbo": true}).Count()

	result_cluster := &ClusterStatus{
		JumboChunksCount: int64(jumbo_chunks),
	}

	result := &MongoStatus{
		ServerStatus:  result_server,
		ReplSetStatus: result_repl,
		ClusterStatus: result_cluster,
	}

	defer func() {
		s.lastResult = result
	}()

	result.SampleTime = time.Now()
	if s.lastResult != nil && result != nil {
		duration := result.SampleTime.Sub(s.lastResult.SampleTime)
		durationInSeconds := int64(duration.Seconds())
		if durationInSeconds == 0 {
			durationInSeconds = 1
		}
		data := NewMongodbData(
			NewStatLine(*s.lastResult, *result, s.Url.Host, true, durationInSeconds),
			s.getDefaultTags(),
		)
		data.AddDefaultStats()
		data.flush(acc)
	}
	return nil
}
