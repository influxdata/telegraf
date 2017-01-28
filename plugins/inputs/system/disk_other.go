// +build !linux

package system

type diskInfoCache struct{}

func (s *DiskIOStats) diskInfo(devName string) (map[string]string, error) {
	return nil, nil
}
