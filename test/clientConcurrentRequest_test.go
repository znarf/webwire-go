package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	tmdwg "github.com/qbeon/tmdwg-go"
	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
	"github.com/stretchr/testify/require"
)

// TestClientConcurrentRequest verifies concurrent calling of client.Request
// is properly synchronized and doesn't cause any data race
func TestClientConcurrentRequest(t *testing.T) {
	concurrentAccessors := 16
	finished := tmdwg.NewTimedWaitGroup(concurrentAccessors*2, 2*time.Second)

	// Initialize webwire server
	server := setupServer(
		t,
		&serverImpl{
			onRequest: func(
				_ context.Context,
				_ wwr.Connection,
				_ wwr.Message,
			) (wwr.Payload, error) {
				finished.Progress(1)
				return nil, nil
			},
		},
		wwr.ServerOptions{},
	)

	// Initialize client
	client := newCallbackPoweredClient(
		server.Addr().String(),
		wwrclt.Options{
			DefaultRequestTimeout: 2 * time.Second,
		},
		callbackPoweredClientHooks{},
	)
	defer client.connection.Close()

	require.NoError(t, client.connection.Connect())

	sendRequest := func() {
		defer finished.Progress(1)
		_, err := client.connection.Request(
			context.Background(),
			"sample",
			wwr.NewPayload(wwr.EncodingBinary, []byte("samplepayload")),
		)
		assert.NoError(t, err)
	}

	for i := 0; i < concurrentAccessors; i++ {
		go sendRequest()
	}

	require.NoError(t, finished.Wait(), "Expectation timed out")
}
