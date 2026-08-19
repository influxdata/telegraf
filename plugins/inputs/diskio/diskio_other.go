//go:build !linux

package diskio

type diskInfoCache struct{}

func (*DiskIO) diskInfo(_ string) (map[string]string, error) {
	return nil, nil
}

func (*DiskIO) resolveName(name string) string {
	return name
}

func (*DiskIO) getDeviceWWID(_ string) string {
	return ""
}
