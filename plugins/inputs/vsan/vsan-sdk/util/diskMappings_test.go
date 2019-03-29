package util

import "testing"

func TestDiskMappingsQuery(t *testing.T) {
	type args struct {
		diskUuid string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
		{name: "happy", args: args{diskUuid: "52adc8e9-a89d-1a56-1908-5e24d74652c4"}, want: "naa.55cd2e414d4f23b3", want1: "10.172.47.149"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := DiskMappingsQuery(tt.args.diskUuid)
			if got != tt.want {
				t.Errorf("DiskMappingsQuery() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DiskMappingsQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
