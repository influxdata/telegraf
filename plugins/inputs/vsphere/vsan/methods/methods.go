package methods

import (
	"context"
	"github.com/influxdata/telegraf/plugins/inputs/vsphere/vsan/types"
	"github.com/vmware/govmomi/vim25/soap"
)

func VsanPerfQueryPerf(ctx context.Context, r soap.RoundTripper, req *types.VsanPerfQueryPerf) (*types.VsanPerfQueryPerfResponse, error) {
	var reqBody, resBody types.VsanPerfQueryPerfBody

	reqBody.Req = req

	if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
		return nil, err
	}

	return resBody.Res, nil
}
