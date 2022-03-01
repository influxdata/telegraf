//go:build !linux
// +build !linux

package diskio

type diskInfoCache struct{}

func (d *DiskIO) diskInfo(devName string) (map[string]string, error) {
	return nil, nil
}
