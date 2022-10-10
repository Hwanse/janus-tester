package peer

import (
	"context"
	"errors"
	"fmt"
	"github.com/Hwanse/janus-tester/internal/janus"
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

func PublishSampleFile(ctx context.Context, p *Peer) <-chan struct{} {

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

	iceCtx, iceConnectedCtxCancel := context.WithCancel(ctx)
	var audioEndCtx context.Context

	if haveAudioFile {
		audioEndCtx, err = AttachAudioSample(ctx, iceCtx, peerConnection)
		if err != nil {
			return nil
		}
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

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

	pubReq := &janus.PublishRequest{
		Request: janus.TypePublish,
	}
	offerMap := map[string]interface{}{
		"type":    offer.Type,
		"sdp":     peerConnection.LocalDescription().SDP,
		"trickle": false,
	}

	pubResponse, err := p.Handle.Publish(pubReq, offerMap)
	if err != nil {
		log.Println("failed to publish request : ", err.Error())
		return nil
	}

	err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  pubResponse["sdp"].(string),
	})
	if err != nil {
		panic(err)
	}

	return audioEndCtx.Done()
}

func AttachAudioSample(ctx context.Context, iceCtx context.Context, pc *webrtc.PeerConnection) (context.Context, error) {
	audioEndCtx, audioEndCtxCancel := context.WithCancel(ctx)

	// Create a audio track
	audioTrack, audioTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if audioTrackErr != nil {
		return nil, audioTrackErr
	}

	rtpSender, err := pc.AddTrack(audioTrack)
	if audioTrackErr != nil {
		return nil, err
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

	go ReadAudioFile(iceCtx, audioTrack, audioEndCtxCancel)

	return audioEndCtx, nil
}

func ReadAudioFile(iceCtx context.Context, audioTrack *webrtc.TrackLocalStaticSample, audioEndCtxCancel context.CancelFunc) {
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
	<-iceCtx.Done()

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
		}

		if oggErr != nil {
			log.Println("ogg file error : ", oggErr.Error())
			return
		}

		// The amount of samples is the difference between the last and current timestamp
		sampleCount := float64(pageHeader.GranulePosition - lastGranule)
		lastGranule = pageHeader.GranulePosition
		sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

		if oggErr = audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); oggErr != nil {
			panic(oggErr)
		}
	}
}
