package parse

import (
	"reflect"
	"testing"
)

func TestRemoveSFXDimensions(t *testing.T) {
	type args struct {
		metricDims map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "remove sf_metric from sfx dimensions",
			args: args{
				metricDims: map[string]string{
					"sf_metric":    "",
					"dimensionKey": "dimensionVal",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RemoveSFXDimensions(tt.args.metricDims)
			if _, isIn := tt.args.metricDims["sf_metric"]; isIn {
				t.Errorf("RemoveSFXDimensions() got metricDims %v, but 'sf_metric' shouldn't be in it", tt.args.metricDims)
			}
		})
	}
}

func TestSetPluginDimension(t *testing.T) {
	type args struct {
		metricName string
		metricDims map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				metricName: "metricName",
				metricDims: map[string]string{
					"dimensionKey": "dimensionVal",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if original, in := tt.args.metricDims["plugin"]; !in {
				SetPluginDimension(tt.args.metricName, tt.args.metricDims)
				if val, in := tt.args.metricDims["plugin"]; !in || val != tt.args.metricName {
					t.Errorf("SetPluginDimension() got %v but wanted plugin dimension with value %s", tt.args.metricDims, tt.args.metricName)
				}
			} else {
				SetPluginDimension(tt.args.metricName, tt.args.metricDims)
				if val, in := tt.args.metricDims["plugin"]; !in || val != original {
					t.Errorf("SetPluginDImension() got %v but wanted plugin dimension with value %s", tt.args.metricDims, original)
				}
			}
		})
	}
}

func TestGetMetricName(t *testing.T) {
	type args struct {
		metric string
		field  string
		dims   map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantsfx bool
	}{
		{
			name: "use sf_metric tag as metric name",
			args: args{
				metric: "datapoint",
				field:  "test",
				dims: map[string]string{
					"sf_metric": "sfxtestmetricname",
				},
			},
			want:    "sfxtestmetricname",
			wantsfx: true,
		},
		{
			name: "fields that equal value should not be append to metricname",
			args: args{
				metric: "datapoint",
				field:  "value",
				dims: map[string]string{
					"testDimKey": "testDimVal",
				},
			},
			want:    "datapoint",
			wantsfx: false,
		},
		{
			name: "fields other than 'value' with out sf_metric dim should return measurement.fieldname as metric name",
			args: args{
				metric: "datapoint",
				field:  "test",
				dims: map[string]string{
					"testDimKey": "testDimVal",
				},
			},
			want:    "datapoint.test",
			wantsfx: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetMetricName(tt.args.metric, tt.args.field, tt.args.dims)
			if got != tt.want {
				t.Errorf("GetMetricName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantsfx {
				t.Errorf("GetMetricName() got1 = %v, want %v", got1, tt.wantsfx)
			}
		})
	}
}

func TestExtractProperty(t *testing.T) {
	type args struct {
		name string
		dims map[string]string
	}
	tests := []struct {
		name      string
		args      args
		wantProps map[string]interface{}
		wantErr   bool
	}{
		{
			name: "ensure that sfx host metadata events remap dimension with key 'property' to properties",
			args: args{
				name: "objects.host-meta-data",
				dims: map[string]string{
					"property":     "propertyValue",
					"dimensionKey": "dimensionValue",
				},
			},
			wantProps: map[string]interface{}{
				"property": "propertyValue",
			},
			wantErr: false,
		},
		{
			name: "malformed sfx host metadata event should return an error",
			args: args{
				name: "objects.host-meta-data",
				dims: map[string]string{
					"dimensionKey": "dimensionValue",
				},
			},
			wantProps: map[string]interface{}{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProps, err := ExtractProperty(tt.args.name, tt.args.dims)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotProps, tt.wantProps) {
				t.Errorf("ExtractProperty() = %v, want %v", gotProps, tt.wantProps)
			}
			if _, ok := tt.args.dims["property"]; ok {
				t.Errorf("ExtractProperty() did not remove property from dims %v", tt.args.dims)
			}
		})
	}
}
