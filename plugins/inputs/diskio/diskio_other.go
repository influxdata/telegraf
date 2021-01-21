// +build !linux

package diskio

type diskInfoCache struct{}

func (s *DiskIO) diskInfo(devName string) (map[string]string, error) {
	return nil, nil
}
