package test

import (
	"github.com/Hwanse/janus-tester/internal/peer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_PublisherPeer(t *testing.T) {
	peerDone := peer.PublishPeer()
	assert.NotNil(t, <-peerDone)
}

func Test_Multi_PublisherPeer(t *testing.T) {
	expectResults := make([]<-chan struct{}, 0)
	for i := 0; i < 8; i++ {
		done := peer.PublishPeer()
		expectResults = append(expectResults, done)
	}

	for _, ch := range expectResults {
		<-ch
	}
}
