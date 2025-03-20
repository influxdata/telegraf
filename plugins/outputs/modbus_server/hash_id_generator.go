package modbus_server

import (
	"github.com/influxdata/telegraf"
	"hash/maphash"
	"sort"
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
		taglist = append(taglist,
			&telegraf.Tag{Key: k, Value: v})
	}
	sort.Slice(taglist, func(i, j int) bool { return taglist[i].Key < taglist[j].Key })

	return genID(g.hashSeed, measurement, taglist)
}

func genID(seed maphash.Seed, measurement string, taglist []*telegraf.Tag) uint64 {
	var mh maphash.Hash
	mh.SetSeed(seed)

	mh.WriteString(measurement)
	mh.WriteByte(0)

	for _, tag := range taglist {
		mh.WriteString(tag.Key)
		mh.WriteByte(0)
		mh.WriteString(tag.Value)
		mh.WriteByte(0)
	}
	mh.WriteByte(0)

	return mh.Sum64()
}
