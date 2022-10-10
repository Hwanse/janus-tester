package peer

import (
	"fmt"
	"github.com/Hwanse/janus-tester/internal/janus"
	"github.com/pion/webrtc/v3"
)

func ConnectPeerConnectionAboutPublisher(p *Peer, jsep map[string]interface{}) error {
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
			p.DestroyFunc()
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

	err = p.Handle.SubscribeStart(&req, answerMap)
	if err != nil {
		return err
	}

	return nil
}
