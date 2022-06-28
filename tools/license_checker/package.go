package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/licensecheck"
)

type packageInfo struct {
	name    string
	version string
	license string
	url     string
	spdx    string
}

func (pkg *packageInfo) ToSPDX() {
	pkg.spdx = nameToSPDX[pkg.license]
}

func (pkg *packageInfo) Classify() (float64, error) {
	// Use the cache if any
	if spdxCache != nil {
		if spdx, confidence, valid := spdxCache.Get(pkg); valid {
			debugf("%q found cache entry: %q with confidence %f%%", pkg.name, spdx, confidence)
			if spdx == pkg.spdx {
				return confidence, nil
			}
			return confidence, fmt.Errorf("classification %q does not match", spdx)
		}
	}

	// Download the license text
	source, err := normalizeUrl(pkg.url)
	if err != nil {
		return 0.0, fmt.Errorf("%q is not a valid URL: %w", pkg.url, err)
	}
	debugf("%q downloading from %q", pkg.name, source)

	response, err := http.Get(source)
	if err != nil {
		return 0.0, fmt.Errorf("download from %q failed: %w", source, err)
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		status := response.StatusCode
		return 0.0, fmt.Errorf("download from %q failed %d: %s", source, status, http.StatusText(status))
	}
	defer response.Body.Close()
	text, err := io.ReadAll(response.Body)
	if err != nil {
		return 0.0, fmt.Errorf("reading body failed: %w", err)
	}
	if len(text) < 1 {
		return 0.0, errors.New("empty body")
	}

	// Classify the license text
	coverage := licensecheck.Scan(text)
	if len(coverage.Match) == 0 {
		return coverage.Percent, errors.New("no match found")
	}
	match := coverage.Match[0]
	debugf("%q found match: %q with confidence %f%%", pkg.name, match.ID, coverage.Percent)

	// Use the cache if any
	if spdxCache != nil {
		debugf("%q adding cache entry: %q with confidence %f%%", pkg.name, match.ID, coverage.Percent)
		spdxCache.Add(pkg, match.ID, coverage.Percent)
	}

	if match.ID != pkg.spdx {
		return coverage.Percent, fmt.Errorf("classification %q does not match", match.ID)
	}
	return coverage.Percent, nil
}

func normalizeUrl(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	switch u.Hostname() {
	case "github.com":
		u.Host = "raw.githubusercontent.com"
		var cleaned []string
		for _, p := range strings.Split(u.Path, "/") {
			// Filter out elements
			if p == "blob" {
				continue
			}
			cleaned = append(cleaned, p)
		}
		u.Path = strings.Join(cleaned, "/")
	case "gitlab.com":
		u.Path = strings.Replace(u.Path, "/-/blob/", "/-/raw/", 1)
	case "git.octo.it":
		parts := strings.Split(u.RawQuery, ";")
		for i, p := range parts {
			if p == "a=blob" {
				parts[i] = "a=blob_plain"
				break
			}
		}
		u.RawQuery = strings.Join(parts, ";")
	}

	return u.String(), nil
}
