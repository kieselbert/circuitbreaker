package main

import (
	"context"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	p := newProcess()

	cfg := &config{
		MaxPendingHtlcs:   2,
		BaseSatPerHr:      1,
		RatePpmPerHr:      5,
		ReportingInterval: time.Minute * 1,
	}

	client := newLndclientMock()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exit := make(chan error)
	go func() {
		exit <- p.run(ctx, client, cfg)
	}()

	key := &routerrpc.CircuitKey{
		ChanId: 2,
		HtlcId: 5,
	}
	client.htlcInterceptorRequests <- &routerrpc.ForwardHtlcInterceptRequest{
		IncomingCircuitKey: key,
	}

	resp := <-client.htlcInterceptorResponses
	require.Equal(t, routerrpc.ResolveHoldForwardAction_RESUME, resp.Action)

	client.htlcEvents <- &routerrpc.HtlcEvent{
		EventType:         routerrpc.HtlcEvent_FORWARD,
		IncomingChannelId: key.ChanId,
		IncomingHtlcId:    key.HtlcId,
		Event:             &routerrpc.HtlcEvent_SettleEvent{},
	}

	time.Sleep(time.Second)

	cancel()
	require.NoError(t, <-exit)
}
