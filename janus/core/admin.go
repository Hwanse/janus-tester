package core

import (
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
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

func (gateway *Gateway) GetStatus() (*SuccessMsg, error) {
	req, ch := newAdminRequest("get_status")
	gateway.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *SuccessMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("get_status")
}
