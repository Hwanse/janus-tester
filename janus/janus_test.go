package janus

import (
	"fmt"
	"testing"
)

const (
	JanusHost               = "127.0.0.1"
	JanusWebsocketPort      = "8188"
	JanusAdminWebsocketPort = "7188"
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

func Test_AdminConnect(t *testing.T) {
	adminUrl := fmt.Sprintf("ws://%s:%s/", JanusHost, JanusAdminWebsocketPort)
	adminClient, err := WsAdminConnect(adminUrl)
	if err != nil {
		t.Fail()
		return
	}
	msg, err := adminClient.GetStatus()
	if err != nil {
		t.Fail()
		t.Log(err)
		return
	}

	t.Log(msg)
}
