package ctrlx_datalayer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

func TestSubscription_createRequest(t *testing.T) {
	tests := []struct {
		name         string
		subscription subscription
		id           string
		wantBody     subscriptionRequest
		wantErr      bool
	}{
		{
			name: "Should_Return_Expected_Request",
			subscription: subscription{
				Nodes: []node{
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
			wantBody: subscriptionRequest{
				Properties: subscriptionProperties{
					KeepaliveInterval: 10000,
					Rules: []rule{
						{
							"Sampling",
							sampling{
								SamplingInterval: 100000,
							},
						},
						{
							"Queueing",
							queueing{
								QueueSize: 100,
								Behaviour: "DiscardNewest",
							},
						},
						{
							"DataChangeFilter",
							dataChangeFilter{
								DeadBandValue: 1.12345,
							},
						},
						{
							"ChangeEvents",
							changeEvents{
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
		nodes   []node
		address string
		want    *node
	}{
		{
			name: "Should_Return_Node_Of_Given_Address",
			nodes: []node{
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
			want: &node{
				Name:    "",
				Address: "path/to/node3",
				Tags:    map[string]string{},
			},
		},
		{
			name: "Should_Return_Nil_If_Node_With_Given_Address_Not_Found",
			nodes: []node{
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
			s := &subscription{
				Nodes: tt.nodes,
			}
			require.Equal(t, tt.want, s.node(tt.address))
		})
	}
}

func TestSubscription_addressList(t *testing.T) {
	tests := []struct {
		name  string
		nodes []node
		want  []string
	}{
		{
			name: "Should_Return_AddressArray_Of_All_Nodes",
			nodes: []node{
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
			s := &subscription{
				Nodes: tt.nodes,
			}
			require.Equal(t, tt.want, s.addressList())
		})
	}
}

func TestNode_fieldKey(t *testing.T) {
	tests := []struct {
		name string
		node node
		want string
	}{
		{
			name: "Should_Return_Name_When_Name_Is_Not_Empty",
			node: node{
				Name:    "used",
				Address: "framework/metrics/system/memused-mb",
			},
			want: "used",
		},
		{
			name: "Should_Return_Address_Base_When_Name_Is_Empty_And_Address_Contains_Full_Path",
			node: node{
				Name:    "",
				Address: "framework/metrics/system/memused-mb",
			},
			want: "memused-mb",
		},
		{
			name: "Should_Return_Address_Base_Root_When_Name_Is_Empty_And_Address_Contains_Root_Path",
			node: node{
				Name:    "",
				Address: "root",
			},
			want: "root",
		},
		{
			name: "Should_Return_Empty_When_Name_and_Address_Are_Empty",
			node: node{
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
