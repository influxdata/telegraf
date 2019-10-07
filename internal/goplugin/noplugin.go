// +build !goplugin

package goplugin

import "errors"

func LoadExternalPlugins(rootDir string) error {
	return errors.New("go plugin support is not enabled")
}
