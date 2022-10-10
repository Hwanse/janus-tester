package test

import (
	"context"
	"fmt"
	"github.com/Hwanse/janus-tester/internal"
	"github.com/Hwanse/janus-tester/internal/janus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_JoinClient(t *testing.T) {
	url := fmt.Sprintf("ws://%s:%s/", janus.JanusLocalHost, janus.JanusWebsocketPort)
	gateway, err := janus.WsConnect(url)
	assert.NoError(t, err)

	session, err := gateway.Create()
	assert.NoError(t, err)

	client := internal.NewClient(session)
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*3)

	client.JoinRoom(ctx, 1234)
	go client.KeepAliveLoop(ctx)

	client.KeepConnection(ctx)
}

func Test_PublishStream(t *testing.T) {
	url := fmt.Sprintf("ws://%s:%s/", janus.JanusLocalHost, janus.JanusWebsocketPort)
	gateway, err := janus.WsConnect(url)
	assert.NoError(t, err)

	session, err := gateway.Create()
	assert.NoError(t, err)

	client := internal.NewClient(session)
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*3)

	client.JoinRoom(ctx, 1234)
	go client.KeepAliveLoop(ctx)

	client.TestPublishStream(ctx)

	client.KeepConnection(ctx)
}
