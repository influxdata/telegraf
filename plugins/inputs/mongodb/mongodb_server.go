package mongodb

import (
	"log"
	"net/url"
	"strings"
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

type oplogEntry struct {
	Timestamp bson.MongoTimestamp `bson:"ts"`
}

func IsAuthorization(err error) bool {
	return strings.Contains(err.Error(), "not authorized")
}

func (s *Server) gatherOplogStats() *OplogStats {
	stats := &OplogStats{}
	localdb := s.Session.DB("local")

	op_first := oplogEntry{}
	op_last := oplogEntry{}
	query := bson.M{"ts": bson.M{"$exists": true}}

	for _, collection_name := range []string{"oplog.rs", "oplog.$main"} {
		if err := localdb.C(collection_name).Find(query).Sort("$natural").Limit(1).One(&op_first); err != nil {
			if err == mgo.ErrNotFound {
				continue
			}
			if IsAuthorization(err) {
				log.Println("D! Error getting first oplog entry (" + err.Error() + ")")
			} else {
				log.Println("E! Error getting first oplog entry (" + err.Error() + ")")
			}
			return stats
		}
		if err := localdb.C(collection_name).Find(query).Sort("-$natural").Limit(1).One(&op_last); err != nil {
			if err == mgo.ErrNotFound || IsAuthorization(err) {
				continue
			}
			if IsAuthorization(err) {
				log.Println("D! Error getting first oplog entry (" + err.Error() + ")")
			} else {
				log.Println("E! Error getting first oplog entry (" + err.Error() + ")")
			}
			return stats
		}
	}

	op_first_time := time.Unix(int64(op_first.Timestamp>>32), 0)
	op_last_time := time.Unix(int64(op_last.Timestamp>>32), 0)
	stats.TimeDiff = int64(op_last_time.Sub(op_first_time).Seconds())
	return stats
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

	resultShards := &ShardStats{}
	err = s.Session.DB("admin").Run(bson.D{
		{
			Name:  "shardConnPoolStats",
			Value: 1,
		},
	}, &resultShards)
	if err != nil {
		if IsAuthorization(err) {
			log.Println("D! Error getting database shard stats (" + err.Error() + ")")
		} else {
			log.Println("E! Error getting database shard stats (" + err.Error() + ")")
		}
	}

	oplogStats := s.gatherOplogStats()

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
		ShardStats:    resultShards,
		OplogStats:    oplogStats,
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
		data.AddShardHostStats()
		data.flush(acc)
	}
	return nil
}
