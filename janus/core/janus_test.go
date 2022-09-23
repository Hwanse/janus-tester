package core

import (
	"fmt"
	"testing"
)

const (
	JanusHost          = "localhost"
	JanusWebsocketPort = "8188"
)

func Test_Connect(t *testing.T) {
	url := fmt.Sprintf("ws://%s:%s/", JanusHost, JanusWebsocketPort)
	client, err := WsConnect(url)
	if err != nil {
		t.Fail()
		return
	}
	mess, err := client.Info()
	if err != nil {
		t.Fail()
		return
	}
	t.Log(mess)

	sess, err := client.Create()
	if err != nil {
		t.Fail()
		return
	}
	t.Log(sess)
	t.Log("connect")
}
