//go:build !linux

package diskio

type diskInfoCache struct{}

func (d *DiskIO) diskInfo(devName string) (map[string]string, error) {
	return nil, nil
}

func resolveName(name string) string {
	return name
}

func getDeviceWWID(name string) string {
	return ""
}
