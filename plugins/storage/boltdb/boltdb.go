package boltdb

import (
	"fmt"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/storage"
	"github.com/ugorji/go/codec"
	bolt "go.etcd.io/bbolt"
)

var (
	codecHandle = &codec.MsgpackHandle{}
)

type BoltDBStorage struct {
	Filename string `toml:"file"`

	db *bolt.DB
}

func (s *BoltDBStorage) Init() error {
	if len(s.Filename) == 0 {
		return fmt.Errorf("Storage service requires filename of db")
	}
	db, err := bolt.Open(s.Filename, 0600, nil)
	if err != nil {
		return fmt.Errorf("couldn't open file %q: %w", s.Filename, err)
	}
	s.db = db
	return nil
}

func (s *BoltDBStorage) Close() error {
	return s.db.Close()
}

func (s *BoltDBStorage) Load(namespace, key string, obj interface{}) error {
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(namespace))
		if b == nil {
			// don't error on not found
			return nil
		}
		v := b.Get([]byte(key))
		decoder := codec.NewDecoderBytes(v, codecHandle)

		if err := decoder.Decode(obj); err != nil {
			return fmt.Errorf("decoding: %w", err)
		}

		return nil
	})
	return err
}

func (s *BoltDBStorage) Save(namespace, key string, value interface{}) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(namespace))
		if b == nil {
			bucket, err := tx.CreateBucket([]byte(namespace))
			if err != nil {
				return err
			}
			b = bucket
		}
		var byt []byte
		enc := codec.NewEncoderBytes(&byt, codecHandle)
		if err := enc.Encode(value); err != nil {
			return fmt.Errorf("encoding: %w", err)
		}
		return b.Put([]byte(key), byt)
	})
}

func (s *BoltDBStorage) GetName() string {
	return "internal"
}

func init() {
	storage.Add("internal", func() config.StoragePlugin {
		return &BoltDBStorage{}
	})
}
