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

func (s *Server) gatherData(acc telegraf.Accumulator, gatherDbStats bool) error {
	s.Session.SetMode(mgo.Eventual, true)
	s.Session.SetSocketTimeout(0)
	result_server := &ServerStatus{}
	err := s.Session.DB("admin").Run(bson.D{
		{
			Name:  "serverStatus",
			Value: 1,
		},
		{
			Name:  "recordStats",
			Value: 0,
		},
	}, result_server)
	if err != nil {
		return err
	}
	result_repl := &ReplSetStatus{}
	// ignore error because it simply indicates that the db is not a member
	// in a replica set, which is fine.
	_ = s.Session.DB("admin").Run(bson.D{
		{
			Name:  "replSetGetStatus",
			Value: 1,
		},
	}, result_repl)

	jumbo_chunks, _ := s.Session.DB("config").C("chunks").Find(bson.M{"jumbo": true}).Count()

	result_cluster := &ClusterStatus{
		JumboChunksCount: int64(jumbo_chunks),
	}

	result_db_stats := &DbStats{}

	if gatherDbStats == true {
		names := []string{}
		names, err = s.Session.DatabaseNames()
		if err != nil {
			log.Println("E! Error getting database names (" + err.Error() + ")")
		}
		for _, db_name := range names {
			db_stat_line := &DbStatsData{}
			err = s.Session.DB(db_name).Run(bson.D{
				{
					Name:  "dbStats",
					Value: 1,
				},
			}, db_stat_line)
			if err != nil {
				log.Println("E! Error getting db stats from " + db_name + "(" + err.Error() + ")")
			}
			db := &Db{
				Name:        db_name,
				DbStatsData: db_stat_line,
			}

			result_db_stats.Dbs = append(result_db_stats.Dbs, *db)
		}
	}

	result := &MongoStatus{
		ServerStatus:  result_server,
		ReplSetStatus: result_repl,
		ClusterStatus: result_cluster,
		DbStats:       result_db_stats,
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
		data.AddDbStats()
		data.flush(acc)
	}
	return nil
}
