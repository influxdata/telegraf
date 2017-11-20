package burrow

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gobwas/glob"
)

func extendUrlPath(src *url.URL, parts ...string) *url.URL {
	dst := new(url.URL)
	*dst = *src

	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}

	ext := strings.Join(parts, "/")
	dst.Path = fmt.Sprintf("%s/%s", src.Path, ext)
	return dst
}

func remapStatus(src string) int {
	switch src {
	case "OK":
		return 1
	case "NOT_FOUND":
		return 2
	case "WARN":
		return 3
	case "ERR":
		return 4
	case "STOP":
		return 5
	case "STALL":
		return 6
	default:
		return 0
	}
}

func makeGlobs(src []string) ([]glob.Glob, error) {
	var dst []glob.Glob
	for _, s := range src {
		g, err := glob.Compile(s)
		if err != nil {
			return nil, err
		}
		dst = append(dst, g)
	}

	return dst, nil
}

func isAllowed(s string, globList []glob.Glob) bool {
	if len(globList) == 0 {
		return true
	}
	for _, g := range globList {
		if g.Match(s) {
			return true
		}
	}
	return false
}
