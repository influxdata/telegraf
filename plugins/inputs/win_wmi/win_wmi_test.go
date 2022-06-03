package win_wmi

import (
	"testing"

	"github.com/influxdata/telegraf"
)

func TestWmi_Gather(t *testing.T) {
	type fields struct {
		Namespace      string
		ClassName      string
		Properties     []string
		Filter         string
		ExcludeNameKey bool
	}
	type args struct {
		acc telegraf.Accumulator
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Initializing COM interface fails in Linux",
			fields:  fields{},
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Wmi{
				Namespace:      tt.fields.Namespace,
				ClassName:      tt.fields.ClassName,
				Properties:     tt.fields.Properties,
				Filter:         tt.fields.Filter,
				ExcludeNameKey: tt.fields.ExcludeNameKey,
			}
			if err := s.Gather(tt.args.acc); (err != nil) != tt.wantErr {
				t.Errorf("Wmi.Gather() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
