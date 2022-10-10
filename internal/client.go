package internal

import (
	"context"
	"github.com/Hwanse/janus-tester/internal/janus"
	"github.com/Hwanse/janus-tester/internal/peer"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
)

type Client struct {
	*janus.Session
	Peers []peer.Peer
}

func NewClient(session *janus.Session) Client {
	return Client{
		session,
		make([]peer.Peer, 0),
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
	c.Peers = append(c.Peers, peer)

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

				if p.MyFeedID != event.Publishers[0].FeedID {
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
	c.WatchRoomEvent(ctx, pubPeer)

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
