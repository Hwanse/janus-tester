package internal

import (
	"context"
	"fmt"
	"github.com/Hwanse/janus-tester/internal/janus"
	"github.com/Hwanse/janus-tester/internal/peer"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
)

type Client struct {
	*janus.Session
	Peers []*peer.Peer
}

func NewClient(session *janus.Session) Client {
	return Client{
		session,
		make([]*peer.Peer, 0),
	}
}

func (c *Client) NewPeer(ctx context.Context, roomID uint64, peerType string) (*peer.Peer, error) {
	_, cancel := context.WithCancel(ctx)
	handle, err := c.Session.Attach(janus.VideoRoomPluginName)
	if err != nil {
		panic(err)
	}

	peer := peer.Peer{
		EnteredRoomID: roomID,
		PeerType:      peerType,
		Handle:        handle,
		DestroyFunc:   cancel,
	}
	c.Peers = append(c.Peers, &peer)

	return &peer, nil
}

func (c *Client) KeepAliveLoop(ctx context.Context) {
	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if _, err := c.Session.KeepAlive(); err != nil {
				log.Println("failed to session keepalive : ", err.Error())
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) WatchRoomEvent(ctx context.Context, p *peer.Peer) {
	// wait for event
	for {
		select {
		case <-ctx.Done():
			return

		default:
			msg := <-p.Handle.Events
			switch msg := msg.(type) {
			case *janus.SlowLinkMsg:
				log.Println("SlowLinkMsg type ", p.Handle.ID)
			case *janus.MediaMsg:
				log.Println("MediaEvent type", msg.Type, " receiving ", msg.Receiving)
			case *janus.WebRTCUpMsg:
				log.Println("WebRTCUp type ", p.Handle.ID)
			case *janus.HangupMsg:
				log.Println("HangupEvent type ", p.Handle.ID)
			case *janus.EventMsg:
				log.Printf("EventMsg %+v", msg.Plugindata.Data)
				event := janus.NewPublisherEvent{}
				err := mapstructure.Decode(msg.Plugindata.Data, &event)
				if err != nil {
					log.Printf("parse event error : %s \n", err.Error())
					continue
				}

				if len(event.Publishers) > 0 && p.MyFeedID != event.Publishers[0].FeedID {
					subPeer, err := c.NewPeer(ctx, event.RoomID, janus.TypeSubscriber)
					if err != nil {
						log.Panic("failed to Create Peer ", err.Error())
						continue
					}

					err = subPeer.SubscribeToPublisher(event.Publishers[0].FeedID)
					if err != nil {
						log.Panic("failed to Create Peer ", err.Error())
						continue
					}
				}
			}
		}
	}
}

func (c *Client) JoinRoom(ctx context.Context, roomID uint64) {
	pubPeer, err := c.NewPeer(ctx, roomID, janus.TypePublisher)
	if err != nil {
		panic(err)
	}
	go c.WatchRoomEvent(ctx, pubPeer)

	joinReq := &janus.JoinPublisherRequest{
		Request:  janus.TypeJoin,
		RoomID:   roomID,
		PeerType: pubPeer.PeerType,
	}

	joinResp, err := pubPeer.Handle.JoinPublisher(joinReq)
	if err != nil {
		panic(err)
	}
	pubPeer.SetMyFeedID(joinResp.FeedID)

	for _, pub := range joinResp.Publishers {
		subPeer, err := c.NewPeer(ctx, roomID, janus.TypeSubscriber)
		if err != nil {
			log.Panic("failed to Create Peer ", err.Error())
			continue
		}

		subPeer.SubscribeToPublisher(pub.FeedID)
	}
}

func (c *Client) LeaveRoom() {
	for _, peer := range c.Peers {
		HandleLeavePeer(peer)
		peer.Handle.Detach()
	}
}

func HandleLeavePeer(peer *peer.Peer) {
	var err error

	switch peer.PeerType {
	case janus.TypePublisher:
		err = peer.Handle.LeavePublisher(&janus.LeaveRequest{Request: janus.TypeLeave})
	case janus.TypeSubscriber:
		err = peer.Handle.LeaveSubscriber(&janus.LeaveRequest{Request: janus.TypeLeave})
	}

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (c *Client) TestPublishStream(ctx context.Context) {
	p := c.FindMyPublisherPeer()
	peer.PublishSampleFile(ctx, p)
}

func (c *Client) KeepConnection(ctx context.Context) {
	for range ctx.Done() {
		log.Println("client connection closed")
		return
	}
}

func (c *Client) FindMyPublisherPeer() *peer.Peer {
	for _, p := range c.Peers {
		if p.PeerType == janus.TypePublisher {
			return p
		}
	}

	return nil
}
