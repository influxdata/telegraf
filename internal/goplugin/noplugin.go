//go:build !goplugin
// +build !goplugin

package goplugin

import "errors"

func LoadExternalPlugins(_ string) error {
	return errors.New("go plugin support is not enabled")
}
