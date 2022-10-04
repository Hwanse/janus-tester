package main

import (
	"context"
	"errors"
	"fmt"
	janus "github.com/Hwanse/janus-tester/internal/janus"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	"io"
	"log"
	"os"
	"time"
)

const (
	audioFileName   = "output.ogg"
	oggPageDuration = time.Millisecond * 20
)

func main() {

	fileInfo, err := os.Stat(audioFileName)
	haveAudioFile := !os.IsNotExist(err)

	if !haveAudioFile {
		panic("Could not find `" + audioFileName + "`")
	}

	log.Println("find audio file : ", fileInfo.Name())

	// Prepare the configuration
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
		panic(err)
	}

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())
	audioEndCtx, audioEndCtxCancel := context.WithCancel(context.Background())

	if haveAudioFile {
		// Create a audio track
		audioTrack, audioTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
		if audioTrackErr != nil {
			panic(audioTrackErr)
		}

		rtpSender, audioTrackErr := peerConnection.AddTrack(audioTrack)
		if audioTrackErr != nil {
			panic(audioTrackErr)
		}

		// Read incoming RTCP packets
		// Before these packets are returned they are processed by interceptors. For things
		// like NACK this needs to be called.
		go func() {
			rtcpBuf := make([]byte, 1500)
			for {
				if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
					return
				}
			}
		}()

		go func() {
			// Open a OGG file and start reading using our OGGReader
			file, oggErr := os.Open(audioFileName)
			if oggErr != nil {
				panic(oggErr)
			}

			// Open on oggfile in non-checksum mode.
			ogg, _, oggErr := oggreader.NewWith(file)
			if oggErr != nil {
				panic(oggErr)
			}

			// Wait for connection established
			<-iceConnectedCtx.Done()

			// Keep track of last granule, the difference is the amount of samples in the buffer
			var lastGranule uint64

			// It is important to use a time.Ticker instead of time.Sleep because
			// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
			// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
			ticker := time.NewTicker(oggPageDuration)
			for ; true; <-ticker.C {
				pageData, pageHeader, oggErr := ogg.ParseNextPage()
				if errors.Is(oggErr, io.EOF) {
					fmt.Printf("All audio pages parsed and sent")
					audioEndCtxCancel()
					os.Exit(0)
				}

				if oggErr != nil {
					panic(oggErr)
				}

				// The amount of samples is the difference between the last and current timestamp
				sampleCount := float64(pageHeader.GranulePosition - lastGranule)
				lastGranule = pageHeader.GranulePosition
				sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

				if oggErr = audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); oggErr != nil {
					panic(oggErr)
				}
			}
		}()
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	// Create a audio track
	opusTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion")
	if err != nil {
		panic(err)
	} else if _, err = peerConnection.AddTrack(opusTrack); err != nil {
		panic(err)
	}

	// Create Offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

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

	go sessionKeepAliveLoop(audioEndCtx, session)
	go watchHandle(audioEndCtx, handle)

	joinReq := &janus.JoinPublisherRequest{
		Request:  janus.TypeJoin,
		RoomID:   1234,
		PeerType: janus.TypePublisher,
	}
	_, err = handle.JoinPublisher(joinReq)
	if err != nil {
		panic(err)
	}

	pubReq := &janus.PublishRequest{
		Request: janus.TypePublish,
	}
	offerMap := map[string]interface{}{
		"type":    "offer",
		"sdp":     peerConnection.LocalDescription().SDP,
		"trickle": false,
	}

	pubResponse, err := handle.Publish(pubReq, offerMap)
	if err != nil {
		log.Println("failed to publish request : ", err.Error())
		return
	}

	if pubResponse != nil {
		err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  pubResponse["sdp"].(string),
		})
		if err != nil {
			panic(err)
		}

	}

	select {}
}

func sessionKeepAliveLoop(ctx context.Context, session *janus.Session) {
	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if _, err := session.KeepAlive(); err != nil {
				log.Println("failed to session keepalive : ", err.Error())
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func watchHandle(ctx context.Context, handle *janus.Handle) {
	// wait for event
	for {
		select {
		case <-ctx.Done():
			return

		default:
			msg := <-handle.Events
			switch msg := msg.(type) {
			case *janus.SlowLinkMsg:
				log.Println("SlowLinkMsg type ", handle.ID)
			case *janus.MediaMsg:
				log.Println("MediaEvent type", msg.Type, " receiving ", msg.Receiving)
			case *janus.WebRTCUpMsg:
				log.Println("WebRTCUp type ", handle.ID)
			case *janus.HangupMsg:
				log.Println("HangupEvent type ", handle.ID)
			case *janus.EventMsg:
				log.Printf("EventMsg %+v", msg.Plugindata.Data)
			}
		}
	}
}
