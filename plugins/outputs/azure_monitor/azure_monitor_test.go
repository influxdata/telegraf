package azure_monitor

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

func TestAzureMonitor_Connect(t *testing.T) {
	type fields struct {
		Timeout             internal.Duration
		NamespacePrefix     string
		StringsAsDimensions bool
		Region              string
		ResourceID          string
		url                 string
		auth                autorest.Authorizer
		client              *http.Client
		cache               map[time.Time]map[uint64]*aggregate
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureMonitor{
				Timeout:             tt.fields.Timeout,
				NamespacePrefix:     tt.fields.NamespacePrefix,
				StringsAsDimensions: tt.fields.StringsAsDimensions,
				Region:              tt.fields.Region,
				ResourceID:          tt.fields.ResourceID,
				url:                 tt.fields.url,
				auth:                tt.fields.auth,
				client:              tt.fields.client,
				cache:               tt.fields.cache,
			}
			if err := a.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("AzureMonitor.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_vmInstanceMetadata(t *testing.T) {
	type args struct {
		c *http.Client
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := vmInstanceMetadata(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("vmInstanceMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("vmInstanceMetadata() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("vmInstanceMetadata() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestAzureMonitor_Write(t *testing.T) {
	type fields struct {
		Timeout             internal.Duration
		NamespacePrefix     string
		StringsAsDimensions bool
		Region              string
		ResourceID          string
		url                 string
		auth                autorest.Authorizer
		client              *http.Client
		cache               map[time.Time]map[uint64]*aggregate
	}
	type args struct {
		metrics []telegraf.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureMonitor{
				Timeout:             tt.fields.Timeout,
				NamespacePrefix:     tt.fields.NamespacePrefix,
				StringsAsDimensions: tt.fields.StringsAsDimensions,
				Region:              tt.fields.Region,
				ResourceID:          tt.fields.ResourceID,
				url:                 tt.fields.url,
				auth:                tt.fields.auth,
				client:              tt.fields.client,
				cache:               tt.fields.cache,
			}
			if err := a.Write(tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("AzureMonitor.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureMonitor_Add(t *testing.T) {
	type fields struct {
		Timeout             internal.Duration
		NamespacePrefix     string
		StringsAsDimensions bool
		Region              string
		ResourceID          string
		url                 string
		auth                autorest.Authorizer
		client              *http.Client
		cache               map[time.Time]map[uint64]*aggregate
	}
	type args struct {
		m telegraf.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureMonitor{
				Timeout:             tt.fields.Timeout,
				NamespacePrefix:     tt.fields.NamespacePrefix,
				StringsAsDimensions: tt.fields.StringsAsDimensions,
				Region:              tt.fields.Region,
				ResourceID:          tt.fields.ResourceID,
				url:                 tt.fields.url,
				auth:                tt.fields.auth,
				client:              tt.fields.client,
				cache:               tt.fields.cache,
			}
			a.Add(tt.args.m)
		})
	}
}

func TestAzureMonitor_Push(t *testing.T) {
	type fields struct {
		Timeout             internal.Duration
		NamespacePrefix     string
		StringsAsDimensions bool
		Region              string
		ResourceID          string
		url                 string
		auth                autorest.Authorizer
		client              *http.Client
		cache               map[time.Time]map[uint64]*aggregate
	}
	tests := []struct {
		name   string
		fields fields
		want   []telegraf.Metric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureMonitor{
				Timeout:             tt.fields.Timeout,
				NamespacePrefix:     tt.fields.NamespacePrefix,
				StringsAsDimensions: tt.fields.StringsAsDimensions,
				Region:              tt.fields.Region,
				ResourceID:          tt.fields.ResourceID,
				url:                 tt.fields.url,
				auth:                tt.fields.auth,
				client:              tt.fields.client,
				cache:               tt.fields.cache,
			}
			if got := a.Push(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AzureMonitor.Push() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureMonitor_Reset(t *testing.T) {
	type fields struct {
		Timeout             internal.Duration
		NamespacePrefix     string
		StringsAsDimensions bool
		Region              string
		ResourceID          string
		url                 string
		auth                autorest.Authorizer
		client              *http.Client
		cache               map[time.Time]map[uint64]*aggregate
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureMonitor{
				Timeout:             tt.fields.Timeout,
				NamespacePrefix:     tt.fields.NamespacePrefix,
				StringsAsDimensions: tt.fields.StringsAsDimensions,
				Region:              tt.fields.Region,
				ResourceID:          tt.fields.ResourceID,
				url:                 tt.fields.url,
				auth:                tt.fields.auth,
				client:              tt.fields.client,
				cache:               tt.fields.cache,
			}
			a.Reset()
		})
	}
}
