package azure_monitor

import (
	"fmt"
	"time"
)

type azureMonitorMetric struct {
	Time  time.Time         `json:"time"`
	Data  *azureMonitorData `json:"data"`
	index int
}

type azureMonitorData struct {
	BaseData *azureMonitorBaseData `json:"baseData"`
}

type azureMonitorBaseData struct {
	Metric         string                `json:"metric"`
	Namespace      string                `json:"namespace"`
	DimensionNames []string              `json:"dimNames"`
	Series         []*azureMonitorSeries `json:"series"`
}

type azureMonitorSeries struct {
	DimensionValues []string `json:"dimValues"`
	Min             float64  `json:"min"`
	Max             float64  `json:"max"`
	Sum             float64  `json:"sum"`
	Count           int64    `json:"count"`
}

// VirtualMachineMetadata contains information about a VM from the metadata service
type virtualMachineMetadata struct {
	Compute struct {
		Location          string `json:"location"`
		Name              string `json:"name"`
		ResourceGroupName string `json:"resourceGroupName"`
		SubscriptionID    string `json:"subscriptionId"`
		VMScaleSetName    string `json:"vmScaleSetName"`
	} `json:"compute"`
}

func (m *virtualMachineMetadata) ResourceID() string {
	if m.Compute.VMScaleSetName != "" {
		return fmt.Sprintf(
			resourceIDScaleSetTemplate,
			m.Compute.SubscriptionID,
			m.Compute.ResourceGroupName,
			m.Compute.VMScaleSetName,
		)
	}

	return fmt.Sprintf(
		resourceIDTemplate,
		m.Compute.SubscriptionID,
		m.Compute.ResourceGroupName,
		m.Compute.Name,
	)
}
