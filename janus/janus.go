// Package core is a Golang implementation of the Janus API, used to interact
// with the Janus WebRTC Gateway.
package janus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

func unexpected(request string) error {
	return fmt.Errorf("unexpected response received to '%s' request", request)
}

func newRequest(method string) (map[string]interface{}, chan interface{}) {
	req := make(map[string]interface{}, 8)
	req["janus"] = method
	return req, make(chan interface{})
}

func newAdminRequest(method string) (map[string]interface{}, chan interface{}) {
	req := make(map[string]interface{}, 8)
	req["janus"] = method
	req["admin_secret"] = AdminSecret
	return req, make(chan interface{})
}

// Gateway represents a connection to an instance of the Janus Gateway.
type Gateway struct {
	// Sessions is a map of the currently active sessions to the gateway.
	Sessions map[uint64]*Session

	// Access to the Sessions map should be synchronized with the Gateway.Lock()
	// and Gateway.Unlock() methods provided by the embeded sync.Mutex.
	sync.Mutex

	conn             *websocket.Conn
	transactions     map[xid.ID]chan interface{}
	transactionsUsed map[xid.ID]bool
	errors           chan error
	sendChan         chan []byte
	writeMu          sync.Mutex
	debug            bool
}

const (
	WebsocketSubProtocol      = "janus-protocol"
	WebsocketAdminSubProtocol = "janus-admin-protocol"
	AdminSecret               = "janusoverlord"
)

func generateTransactionId() xid.ID {
	return xid.New()
}

// WsConnect initiates a websocket connection with the Janus Gateway
func WsConnect(wsURL string) (*Gateway, error) {
	websocket.DefaultDialer.Subprotocols = []string{WebsocketSubProtocol}

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

// Close closes the underlying connection to the Gateway.
func (gateway *Gateway) Close() error {
	return gateway.conn.Close()
}

// GetErrChan returns a channels through which the caller can check and react to connectivity errors
func (gateway *Gateway) GetErrChan() chan error {
	return gateway.errors
}

func (gateway *Gateway) send(msg map[string]interface{}, transaction chan interface{}) {
	guid := generateTransactionId()

	msg["transaction"] = guid.String()
	gateway.Lock()
	gateway.transactions[guid] = transaction
	gateway.transactionsUsed[guid] = false
	gateway.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("json.Marshal: %s\n", err)
		return
	}

	if gateway.debug {
		// log message being sent
		var log bytes.Buffer
		json.Indent(&log, data, ">", "   ")
		log.Write([]byte("\n"))
		log.WriteTo(os.Stdout)
	}

	gateway.writeMu.Lock()
	err = gateway.conn.WriteMessage(websocket.TextMessage, data)
	gateway.writeMu.Unlock()

	if err != nil {
		select {
		case gateway.errors <- err:
		default:
			fmt.Printf("conn.Write: %s\n", err)
		}

		return
	}
}

func passMsg(ch chan interface{}, msg interface{}) {
	ch <- msg
}

func (gateway *Gateway) ping() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := gateway.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(20*time.Second))
			if err != nil {
				select {
				case gateway.errors <- err:
				default:
					log.Println("ping:", err)
				}

				return
			}
		}
	}
}

func (gateway *Gateway) sendloop() {

}

func (gateway *Gateway) recv() {

	for {
		// Read message from Gateway

		// Decode to Msg struct
		var base BaseMsg

		_, data, err := gateway.conn.ReadMessage()
		if err != nil {
			select {
			case gateway.errors <- err:
			default:
				fmt.Printf("conn.Read: %s\n", err)
			}

			return
		}

		if err := json.Unmarshal(data, &base); err != nil {
			fmt.Printf("json.Unmarshal: %s\n", err)
			continue
		}

		if gateway.debug {
			// log message being sent
			var log bytes.Buffer
			json.Indent(&log, data, "<", "   ")
			log.Write([]byte("\n"))
			log.WriteTo(os.Stdout)
		}

		typeFunc, ok := msgtypes[base.Type]
		if !ok {
			fmt.Printf("Unknown message type received!\n")
			continue
		}

		msg := typeFunc()
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Printf("json.Unmarshal: %s\n", err)
			continue // Decode error
		}
		ifSuccessMsgAppendJsonData(msg, data)

		var transactionUsed bool
		if base.ID != "" {
			id, _ := xid.FromString(base.ID)
			gateway.Lock()
			transactionUsed = gateway.transactionsUsed[id]
			gateway.Unlock()

		}

		// Pass message on from here
		if base.ID == "" || transactionUsed {
			// Is this a Handle event?
			if base.Handle == 0 {
				// Error()
			} else {
				// Lookup Session
				gateway.Lock()
				session := gateway.Sessions[base.Session]
				gateway.Unlock()
				if session == nil {
					fmt.Printf("Unable to deliver message. Session gone?\n")
					continue
				}

				// Lookup Handle
				session.Lock()
				handle := session.Handles[base.Handle]
				session.Unlock()
				if handle == nil {
					fmt.Printf("Unable to deliver message. Handle gone?\n")
					continue
				}

				// Pass msg
				go passMsg(handle.Events, msg)
			}
		} else {
			id, _ := xid.FromString(base.ID)
			// Lookup Transaction
			gateway.Lock()
			transaction := gateway.transactions[id]
			switch msg.(type) {
			case *EventMsg:
				gateway.transactionsUsed[id] = true
			}
			gateway.Unlock()
			if transaction == nil {
				// Error()
			}

			// Pass msg
			go passMsg(transaction, msg)
		}
	}
}

// Info sends an info request to the Gateway.
// On success, an InfoMsg will be returned and error will be nil.
func (gateway *Gateway) Info() (*InfoMsg, error) {
	req, ch := newRequest("info")
	gateway.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *InfoMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("info")
}

// Create sends a create request to the Gateway.
// On success, a new Session will be returned and error will be nil.
func (gateway *Gateway) Create() (*Session, error) {
	req, ch := newRequest("create")
	gateway.send(req, ch)

	msg := <-ch
	var success *SuccessMsg
	switch msg := msg.(type) {
	case *SuccessMsg:
		success = msg
	case *ErrorMsg:
		return nil, msg
	}

	// Create new session
	session := new(Session)
	session.gateway = gateway
	session.ID = success.Data.ID
	session.Handles = make(map[uint64]*Handle)
	session.Events = make(chan interface{}, 2)

	// Store this session
	gateway.Lock()
	gateway.Sessions[session.ID] = session
	gateway.Unlock()

	return session, nil
}

func ifSuccessMsgAppendJsonData(msg interface{}, response []byte) {
	switch msg := msg.(type) {
	case *SuccessMsg:
		msg.response = response
	}
}
