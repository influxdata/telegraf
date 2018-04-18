package signalfxmetadata

import (
	// "reflect"
	"testing"
)

// func TestNewAWSInfo(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want *AWSInfo
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewAWSInfo(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewAWSInfo() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestAWSInfo_GetAWSInfo(t *testing.T) {
// 	type fields struct {
// 		aws         bool
// 		awsSet      bool
// 		awsUniqueID string
// 	}
// 	type args struct {
// 		info map[string]string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		{
// 			name:   "",
// 			fields: fields{
// 				aws: true,
// 				awsSet: false,
// 				awsUniqueID: ""
// 			},
// 			args:   {},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &AWSInfo{
// 				aws:         tt.fields.aws,
// 				awsSet:      tt.fields.awsSet,
// 				awsUniqueID: tt.fields.awsUniqueID,
// 			}
// 			s.GetAWSInfo(tt.args.info)
// 		})
// 	}
// }

func Test_buildAWSUniqueID(t *testing.T) {
	type args struct {
		info map[string]string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "Validate that AWSUniqueID can build",
			args: args{
				info: map[string]string{
					"aws_instance_id": "hello",
					"aws_region":      "world",
					"aws_account_id":  "too",
				},
			},
			want:  "hello_world_too",
			want1: true,
		},
		{
			name: "AWSUniqueID missing aws_instance_id should produce error",
			args: args{
				info: map[string]string{
					"aws_region":     "world",
					"aws_account_id": "too",
				},
			},
			want:  "",
			want1: false,
		},
		{
			name: "AWSUniqueID missing aws_region should produce error",
			args: args{
				info: map[string]string{
					"aws_instance_id": "hello",
					"aws_account_id":  "too",
				},
			},
			want:  "",
			want1: false,
		},
		{
			name: "AWSUniqueID missing aws_account_id should produce error",
			args: args{
				info: map[string]string{
					"aws_region":      "world",
					"aws_instance_id": "hello",
				},
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildAWSUniqueID(tt.args.info)
			if got != tt.want {
				t.Errorf("buildAWSUniqueID() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("buildAWSUniqueID() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

// func Test_processAWSInfo(t *testing.T) {
// 	type args struct {
// 		info     map[string]string
// 		identity map[string]interface{}
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			processAWSInfo(tt.args.info, tt.args.identity)
// 		})
// 	}
// }

// func Test_requestAWSInfo(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		want    map[string]interface{}
// 		wantErr bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := requestAWSInfo()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("requestAWSInfo() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("requestAWSInfo() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
