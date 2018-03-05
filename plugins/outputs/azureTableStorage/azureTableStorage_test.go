package azureTableStorage

import (
	"fmt"
	"io"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/Azure/azure-storage-go"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/testutil"
)

func TestBuildDimensions(t *testing.T) {
	return nil
}