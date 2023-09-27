package ctrlx_datalayer

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/require"
)

func TestSubscription_createRequest(t *testing.T) {
	tests := []struct {
		name         string
		subscription Subscription
		id           string
		wantBody     SubscriptionRequest
		wantErr      bool
	}{
		{
			name: "Should_Return_Expected_Request",
			subscription: Subscription{
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
			id: "sub_id",
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
			got := tt.subscription.createRequest(tt.id)
			require.Equal(t, tt.wantBody, got)
		})
	}
}

func TestSubscription_node(t *testing.T) {
	tests := []struct {
		name    string
		nodes   []Node
		address string
		want    *Node
	}{
		{
			name: "Should_Return_Node_Of_Given_Address",
			nodes: []Node{
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
			address: "path/to/node3",
			want: &Node{
				Name:    "",
				Address: "path/to/node3",
				Tags:    map[string]string{},
			},
		},
		{
			name: "Should_Return_Nil_If_Node_With_Given_Address_Not_Found",
			nodes: []Node{
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
			address: "path/to/node4",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscription{
				Nodes: tt.nodes,
			}
			require.Equal(t, tt.want, s.node(tt.address))
		})
	}
}

func TestSubscription_addressList(t *testing.T) {
	tests := []struct {
		name  string
		nodes []Node
		want  []string
	}{
		{
			name: "Should_Return_AddressArray_Of_All_Nodes",
			nodes: []Node{
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
				Nodes: tt.nodes,
			}
			require.Equal(t, tt.want, s.addressList())
		})
	}
}

func TestNode_fieldKey(t *testing.T) {
	tests := []struct {
		name string
		node Node
		want string
	}{
		{
			name: "Should_Return_Name_When_Name_Is_Not_Empty",
			node: Node{
				Name:    "used",
				Address: "framework/metrics/system/memused-mb",
			},
			want: "used",
		},
		{
			name: "Should_Return_Address_Base_When_Name_Is_Empty_And_Address_Contains_Full_Path",
			node: Node{
				Name:    "",
				Address: "framework/metrics/system/memused-mb",
			},
			want: "memused-mb",
		},
		{
			name: "Should_Return_Address_Base_Root_When_Name_Is_Empty_And_Address_Contains_Root_Path",
			node: Node{
				Name:    "",
				Address: "root",
			},
			want: "root",
		},
		{
			name: "Should_Return_Empty_When_Name_and_Address_Are_Empty",
			node: Node{
				Name:    "",
				Address: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.node.fieldKey())
		})
	}
}
