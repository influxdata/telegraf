package ctrlx_datalayer

import (
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
)

func TestSubscription_createRequest(t *testing.T) {
	type fields struct {
		Nodes             []Node
		Tags              map[string]string
		Measurement       string
		PublishInterval   config.Duration
		KeepaliveInterval config.Duration
		ErrorInterval     config.Duration
		SamplingInterval  config.Duration
		QueueSize         uint
		QueueBehaviour    string
		DeadBandValue     float64
		ValueChange       string
		OutputJSONString  bool
	}
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantBody SubscriptionRequest
		wantErr  bool
	}{
		{
			name: "Should_Return_Expected_Request",
			fields: fields{
				Nodes: []Node{
					{
						Name:    "node1",
						Address: "path/to/node1",
						Tags:    map[string]string{},
					},
					{
						Name:    "node2",
						Address: "path/to/node2",
						Tags:    map[string]string{},
					},
				},
				Tags:              map[string]string{},
				Measurement:       "",
				PublishInterval:   config.Duration(2 * time.Second),
				KeepaliveInterval: config.Duration(10 * time.Second),
				ErrorInterval:     config.Duration(20 * time.Second),
				SamplingInterval:  config.Duration(100 * time.Millisecond),
				QueueSize:         100,
				QueueBehaviour:    "DiscardNewest",
				DeadBandValue:     1.12345,
				ValueChange:       "StatusValueTimestamp",
				OutputJSONString:  true,
			},
			args: args{
				id: "sub_id",
			},
			wantBody: SubscriptionRequest{
				Properties: SubscriptionProperties{
					KeepaliveInterval: 10000,
					Rules: []Rule{
						{
							"Sampling",
							Sampling{
								SamplingInterval: 100000,
							},
						},
						{
							"Queueing",
							Queueing{
								QueueSize: 100,
								Behaviour: "DiscardNewest",
							},
						},
						{
							"DataChangeFilter",
							DataChangeFilter{
								DeadBandValue: 1.12345,
							},
						},
						{
							"ChangeEvents",
							ChangeEvents{
								ValueChange: "StatusValueTimestamp",
							},
						},
					},
					ID:              "sub_id",
					PublishInterval: 2000,
					ErrorInterval:   20000,
				},
				Nodes: []string{
					"path/to/node1",
					"path/to/node2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscription{
				Nodes:             tt.fields.Nodes,
				Tags:              tt.fields.Tags,
				Measurement:       tt.fields.Measurement,
				PublishInterval:   tt.fields.PublishInterval,
				KeepaliveInterval: tt.fields.KeepaliveInterval,
				ErrorInterval:     tt.fields.ErrorInterval,
				SamplingInterval:  tt.fields.SamplingInterval,
				QueueSize:         tt.fields.QueueSize,
				QueueBehaviour:    tt.fields.QueueBehaviour,
				DeadBandValue:     tt.fields.DeadBandValue,
				ValueChange:       tt.fields.ValueChange,
				OutputJSONString:  tt.fields.OutputJSONString,
			}
			got := s.createRequest(tt.args.id)
			require.Equal(t, tt.wantBody, got)
		})
	}
}

func TestSubscription_node(t *testing.T) {
	type fields struct {
		Nodes []Node
	}
	type args struct {
		address string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Node
	}{
		{
			name: "Should_Return_Node_Of_Given_Address",
			fields: fields{
				Nodes: []Node{
					{
						Name:    "node1",
						Address: "path/to/node1",
						Tags:    map[string]string{},
					},
					{
						Name:    "node2",
						Address: "path/to/node2",
						Tags:    map[string]string{},
					},
					{
						Name:    "",
						Address: "path/to/node3",
						Tags:    map[string]string{},
					},
				},
			},
			args: args{
				address: "path/to/node3",
			},
			want: &Node{
				Name:    "",
				Address: "path/to/node3",
				Tags:    map[string]string{},
			},
		},
		{
			name: "Should_Return_Nil_If_Node_With_Given_Address_Not_Found",
			fields: fields{
				Nodes: []Node{
					{
						Name:    "Node1",
						Address: "path/to/node1",
						Tags:    map[string]string{},
					},
					{
						Name:    "Node2",
						Address: "path/to/node2",
						Tags:    map[string]string{},
					},
					{
						Name:    "",
						Address: "path/to/node3",
						Tags:    map[string]string{},
					},
				},
			},
			args: args{
				address: "path/to/node4",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscription{
				Nodes: tt.fields.Nodes,
			}
			if got := s.node(tt.args.address); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscription.node() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscription_addressList(t *testing.T) {
	type fields struct {
		Nodes []Node
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Should_Return_AddressArray_Of_All_Nodes",
			fields: fields{
				Nodes: []Node{
					{
						Address: "framework/metrics/system/memused-mb",
					},
					{
						Address: "framework/metrics/system/memavailable-mb",
					},
					{
						Address: "root",
					},
					{
						Address: "",
					},
				},
			},
			want: []string{
				"framework/metrics/system/memused-mb",
				"framework/metrics/system/memavailable-mb",
				"root",
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscription{
				Nodes: tt.fields.Nodes,
			}
			if got := s.addressList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscription.addressList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_fieldKey(t *testing.T) {
	type fields struct {
		Name    string
		Address string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Should_Return_Name_When_Name_Is_Not_Empty",
			fields: fields{
				Name:    "used",
				Address: "framework/metrics/system/memused-mb",
			},
			want: "used",
		},
		{
			name: "Should_Return_Address_Base_When_Name_Is_Empty_And_Address_Contains_Full_Path",
			fields: fields{
				Name:    "",
				Address: "framework/metrics/system/memused-mb",
			},
			want: "memused-mb",
		},
		{
			name: "Should_Return_Address_Base_Root_When_Name_Is_Empty_And_Address_Contains_Root_Path",
			fields: fields{
				Name:    "",
				Address: "root",
			},
			want: "root",
		},
		{
			name: "Should_Return_Empty_When_Name_and_Address_Are_Empty",
			fields: fields{
				Name:    "",
				Address: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				Name:    tt.fields.Name,
				Address: tt.fields.Address,
			}
			if got := n.fieldKey(); got != tt.want {
				t.Errorf("Node.fieldKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
