package gnmilistener

import (
	"context"

	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/tap"
)

func (g *GNMIListener) logCalls(ctx context.Context, info *tap.Info) (context.Context, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, nil
	}

	agent := "unknown"
	if v := info.Header.Get("user-agent"); len(v) > 0 {
		agent = v[0]
	}
	g.Log.Tracef("%s calling %q (user-agent: %s)", p.Addr.String(), info.FullMethodName, agent)

	return ctx, nil
}
