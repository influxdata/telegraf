package modbus_server

import (
	"hash/maphash"
	"sort"

	"github.com/influxdata/telegraf"
)

func NewHashIDGenerator() *HashIDGenerator {
	return &HashIDGenerator{
		hashSeed: maphash.MakeSeed(),
	}
}

type HashIDGenerator struct {
	hashSeed maphash.Seed
}

func (g *HashIDGenerator) GetID(
	measurement string,
	tags map[string]string,
) uint64 {
	taglist := make([]*telegraf.Tag, 0, len(tags))
	for k, v := range tags {
		taglist = append(
			taglist,
			&telegraf.Tag{Key: k, Value: v},
		)
	}
	sort.Slice(taglist, func(i, j int) bool { return taglist[i].Key < taglist[j].Key })

	return genID(g.hashSeed, measurement, taglist)
}

func genID(seed maphash.Seed, measurement string, taglist []*telegraf.Tag) uint64 {
	var mh maphash.Hash
	mh.SetSeed(seed)

	_, err := mh.WriteString(measurement)
	if err != nil {
		return 0
	}
	err = mh.WriteByte(0)
	if err != nil {
		return 0
	}

	for _, tag := range taglist {
		_, err = mh.WriteString(tag.Key)
		if err != nil {
			return 0
		}
		err = mh.WriteByte(0)
		if err != nil {
			return 0
		}
		_, err = mh.WriteString(tag.Value)
		if err != nil {
			return 0
		}
		err = mh.WriteByte(0)
		if err != nil {
			return 0
		}
	}
	err = mh.WriteByte(0)
	if err != nil {
		return 0
	}

	return mh.Sum64()
}
