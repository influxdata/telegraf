package win_wmi

import (
	"testing"

	ole "github.com/go-ole/go-ole"
	"github.com/influxdata/telegraf"
)

func Test_oleInt64(t *testing.T) {
	type args struct {
		item *ole.IDispatch
		prop string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "fails with no args",
			args:    args{},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := oleInt64(tt.args.item, tt.args.prop)
			if (err != nil) != tt.wantErr {
				t.Errorf("oleInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("oleInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
