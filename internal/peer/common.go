package peer

import (
	"context"
	"github.com/Hwanse/janus-tester/internal/janus"
	"log"
)

type Peer struct {
	MyFeedID      uint64
	EnteredRoomID uint64
	PeerType      string
	Handle        *janus.Handle
	DestroyFunc   context.CancelFunc
}

func (p *Peer) SetMyFeedID(id uint64) {
	p.MyFeedID = id
}

func (p *Peer) SubscribeToPublisher(targetFeedID uint64) error {
	req := janus.JoinSubscriberRequest{
		Request:  janus.TypeJoin,
		RoomID:   p.EnteredRoomID,
		PeerType: p.PeerType,
		Streams:  []janus.Stream{{FeedID: targetFeedID}},
	}
	response, err := p.Handle.JoinSubscriber(&req)
	if err != nil {
		log.Panic("failed to join subscriber : ", err.Error())
		return err
	}

	err = ConnectPeerConnectionAboutPublisher(p, response.Jsep)
	if err != nil {
		log.Panic("failed to join subscriber : ", err.Error())
		return err
	}

	return nil
}
