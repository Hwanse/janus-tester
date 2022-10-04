package janus

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"github.com/tidwall/gjson"
)

func WsAdminConnect(wsURL string) (*Gateway, error) {
	websocket.DefaultDialer.Subprotocols = []string{WebsocketAdminSubProtocol}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	gateway := new(Gateway)
	gateway.conn = conn
	gateway.transactions = make(map[xid.ID]chan interface{})
	gateway.transactionsUsed = make(map[xid.ID]bool)
	gateway.Sessions = make(map[uint64]*Session)
	gateway.sendChan = make(chan []byte, 100)
	gateway.errors = make(chan error)
	gateway.debug = true

	go gateway.ping()
	go gateway.recv()
	return gateway, nil
}

func (gateway *Gateway) GetStatus() (interface{}, error) {
	req, ch := newAdminRequest("get_status")
	gateway.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *SuccessMsg:
		if !gjson.GetBytes(msg.response, "status").Exists() {
			return nil, errors.New("get_status response not contains status field")
		}
		statusInfo := gjson.GetBytes(msg.response, "status").Value()
		return statusInfo, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("get_status")
}
