package internal

import (
	"context"
	"fmt"
	"github.com/Hwanse/janus-tester/internal/janus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Client(t *testing.T) {
	url := fmt.Sprintf("ws://%s:%s/", janus.JanusLocalHost, janus.JanusWebsocketPort)
	gateway, err := janus.WsConnect(url)
	assert.NoError(t, err)

	session, err := gateway.Create()
	assert.NoError(t, err)

	client := NewClient(session)
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)

	client.JoinRoom(ctx, 1234)
	go client.KeepAliveLoop(ctx)

	client.KeepConnection(ctx)
}
