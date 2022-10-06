package peer

import (
	"context"
	"fmt"
	"github.com/Hwanse/janus-tester/internal/janus"
	"github.com/pion/webrtc/v3"
	"log"
)

func SubscriberPeer(cancel context.CancelFunc, handle *janus.Handle, jsep map[string]interface{}) error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			cancel()
		}
	})

	err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  jsep["sdp"].(string),
	})
	if err != nil {
		return err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	req := janus.SubscribeStartRequest{Request: janus.TypeStart}
	answerMap := map[string]interface{}{
		"type":    answer.Type,
		"sdp":     answer.SDP,
		"trickle": false,
	}

	err = handle.SubscribeStart(&req, answerMap)
	if err != nil {
		return err
	}

	return nil
}

func JoinInRoom(roomID uint64) {
	ctx, cancel := context.WithCancel(context.Background())

	janusURL := fmt.Sprintf("ws://%s:%s/", janus.JanusLocalHost, janus.JanusWebsocketPort)
	gateway, err := janus.WsConnect(janusURL)
	if err != nil {
		panic(err)
	}

	session, err := gateway.Create()
	if err != nil {
		panic(err)
	}

	handle, err := session.Attach(janus.VideoRoomPluginName)
	if err != nil {
		panic(err)
	}

	go sessionKeepAliveLoop(ctx, session)
	go watchHandle(ctx, handle)

	joinReq := &janus.JoinPublisherRequest{
		Request:  janus.TypeJoin,
		RoomID:   roomID,
		PeerType: janus.TypePublisher,
	}

	joinResp, err := handle.JoinPublisher(joinReq)
	if err != nil {
		panic(err)
	}

	for _, p := range joinResp.Publishers {
		session, err := gateway.Create()
		if err != nil {
			panic(err)
		}

		handle, err := session.Attach(janus.VideoRoomPluginName)
		if err != nil {
			panic(err)
		}

		go sessionKeepAliveLoop(ctx, session)
		go watchHandle(ctx, handle)

		req := janus.JoinSubscriberRequest{
			Request:  janus.TypeJoin,
			RoomID:   roomID,
			PeerType: janus.TypeSubscriber,
			Streams:  []janus.Stream{{FeedID: p.FeedID}},
		}
		response, err := handle.JoinSubscriber(&req)
		if err != nil {
			log.Panic("failed to join subscriber : ", err.Error())
			break
		}

		err = SubscriberPeer(cancel, handle, response.Jsep)
		if err != nil {
			log.Panic("failed to join subscriber : ", err.Error())
			break
		}
	}

	select {}
}
