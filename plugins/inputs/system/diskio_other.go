// +build !linux

package system

type diskInfoCache struct{}

func (s *DiskIO) diskInfo(devName string) (map[string]string, error) {
	return nil, nil
}
