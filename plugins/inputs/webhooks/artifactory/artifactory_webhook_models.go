package artifactory

import (
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

const meas = "artifactory_webhooks"

type CommonFields struct {
	Domain string `json:"domain"`
	Event  string `json:"event_type"`
}

type Event interface {
	NewMetric() telegraf.Metric
}

type ArtifactDeploymentOrDeletedEvent struct {
	Domain string `json:"domain"`
	Event  string `json:"event_type"`
	Data   struct {
		Repo string `json:"repo_key"`
		Path string `json:"path"`
		Name string `json:"name"`
		Size int64  `json:"size"`
		Sha  string `json:"sha256"`
	} `json:"data"`
}

func (e ArtifactDeploymentOrDeletedEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"domain":     e.Domain,
		"event_type": e.Event,
		"repo":       e.Data.Repo,
		"path":       e.Data.Path,
		"name":       e.Data.Name,
	}
	f := map[string]interface{}{
		"size":   e.Data.Size,
		"sha256": e.Data.Sha,
	}

	return metric.New(meas, t, f, time.Now())
}

type ArtifactMovedOrCopiedEvent struct {
	Domain string `json:"domain"`
	Event  string `json:"event_type"`
	Data   struct {
		Repo       string `json:"repo_key"`
		Path       string `json:"path"`
		Name       string `json:"name"`
		Size       int64  `json:"size"`
		SourcePath string `json:"source_repo_path"`
		TargetPath string `json:"target_repo_path"`
	} `json:"data"`
}

func (e ArtifactMovedOrCopiedEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"domain":     e.Domain,
		"event_type": e.Event,
		"repo":       e.Data.Repo,
		"path":       e.Data.Path,
		"name":       e.Data.Name,
	}
	f := map[string]interface{}{
		"size":        e.Data.Size,
		"source_path": e.Data.SourcePath,
		"target_path": e.Data.TargetPath,
	}

	return metric.New(meas, t, f, time.Now())
}

type ArtifactPropertiesEvent struct {
	Domain string `json:"domain"`
	Event  string `json:"event_type"`
	Data   struct {
		Repo           string   `json:"repo_key"`
		Path           string   `json:"path"`
		Name           string   `json:"name"`
		Size           int64    `json:"size"`
		PropertyKey    string   `json:"property_key"`
		PropertyValues []string `json:"property_values"`
	}
}

func (e ArtifactPropertiesEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"domain":     e.Domain,
		"event_type": e.Event,
		"repo":       e.Data.Repo,
		"path":       e.Data.Path,
		"name":       e.Data.Name,
	}

	f := map[string]interface{}{
		"property_key":    e.Data.PropertyKey,
		"property_values": strings.Join(e.Data.PropertyValues, ","),
	}

	return metric.New(meas, t, f, time.Now())
}

type DockerEvent struct {
	Domain string `json:"domain"`
	Event  string `json:"event_type"`
	Data   struct {
		Repo      string `json:"repo_key"`
		Path      string `json:"path"`
		Name      string `json:"name"`
		Size      int64  `json:"size"`
		Sha       string `json:"sha256"`
		ImageName string `json:"image_name"`
		Tag       string `json:"tag"`
		Platforms []struct {
			Architecture string `json:"achitecture"`
			Os           string `json:"os"`
		} `json:"platforms"`
	} `json:"data"`
}

func (e DockerEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"domain":     e.Domain,
		"event_type": e.Event,
		"repo":       e.Data.Repo,
		"path":       e.Data.Path,
		"name":       e.Data.Name,
		"image_name": e.Data.ImageName,
	}
	f := map[string]interface{}{
		"size":      e.Data.Size,
		"sha256":    e.Data.Sha,
		"tag":       e.Data.Tag,
		"platforms": e.Data.Platforms,
	}

	return metric.New(meas, t, f, time.Now())
}

type BuildEvent struct {
	Domain string `json:"domain"`
	Event  string `json:"event_type"`
	Data   struct {
		BuildName    string `json:"build_name"`
		BuildNumber  string `json:"build_number"`
		BuildStarted string `json:"build_started"`
	} `json:"data"`
}

func (e BuildEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"domain":     e.Domain,
		"event_type": e.Event,
	}
	f := map[string]interface{}{
		"build_name":    e.Data.BuildName,
		"build_number":  e.Data.BuildNumber,
		"build_started": e.Data.BuildStarted,
	}

	return metric.New(meas, t, f, time.Now())
}

type ReleaseBundleEvent struct {
	Domain      string `json:"domain"`
	Event       string `json:"event_type"`
	Destination string `json:"destination"`
	Data        struct {
		ReleaseBundleName    string `json:"release_bundle_name"`
		ReleaseBundleSize    int64  `json:"release_bundle_size"`
		ReleaseBundleVersion string `json:"release_bundle_version"`
	} `json:"data"`
	JpdOrigin string `json:"jpd_origin"`
}

func (e ReleaseBundleEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"domain":              e.Domain,
		"event_type":          e.Event,
		"destination":         e.Destination,
		"release_bundle_name": e.Data.ReleaseBundleName,
	}
	f := map[string]interface{}{
		"release_bundle_size":    e.Data.ReleaseBundleSize,
		"release_bundle_version": e.Data.ReleaseBundleVersion,
		"jpd_origin":             e.JpdOrigin,
	}

	return metric.New(meas, t, f, time.Now())
}

type DistributionEvent struct {
	Domain      string `json:"domain"`
	Event       string `json:"event_type"`
	Destination string `json:"destination"`
	Data        struct {
		EdgeNodeInfoList []struct {
			EdgeNodeAddress string `json:"edge_node_address"`
			EdgeNodeName    string `json:"edge_node_name"`
		} `json:"edge_node_info_list"`
		Name          string `json:"release_bundle_name"`
		Size          int64  `json:"release_bundle_size"`
		Version       string `json:"release_bundle_version"`
		Message       string `json:"status_message"`
		TransactionID int64  `json:"transaction_id"`
	} `json:"data"`
	OriginURL string `json:"jpd_origin"`
}

func (e DistributionEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"domain":              e.Domain,
		"event_type":          e.Event,
		"destination":         e.Destination,
		"release_bundle_name": e.Data.Name,
	}
	f := map[string]interface{}{
		"release_bundle_size":    e.Data.Size,
		"release_bundle_version": e.Data.Version,
		"status_message":         e.Data.Message,
		"transaction_id":         e.Data.TransactionID,
		"edge_node_info_list":    e.Data.EdgeNodeInfoList,
		"jpd_origin":             e.OriginURL,
	}
	return metric.New(meas, t, f, time.Now())
}

type DestinationEvent struct {
	Domain      string `json:"domain"`
	Event       string `json:"event_type"`
	Destination string `json:"destination"`
	Data        struct {
		Name    string `json:"release_bundle_name"`
		Version string `json:"release_bundle_version"`
		Message string `json:"status_message"`
	} `json:"data"`
	OriginURL string `json:"jpd_origin"`
}

func (e DestinationEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"domain":              e.Domain,
		"event_type":          e.Event,
		"destination":         e.Destination,
		"release_bundle_name": e.Data.Name,
	}
	f := map[string]interface{}{
		"release_bundle_version": e.Data.Version,
		"status_message":         e.Data.Message,
		"jpd_origin":             e.OriginURL,
	}
	return metric.New(meas, t, f, time.Now())
}
