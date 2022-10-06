package test

import (
	"github.com/Hwanse/janus-tester/internal/peer"
	"testing"
	"time"
)

func Test_Subscribe(t *testing.T) {
	for i := 0; i < 10; i++ {
		go peer.JoinInRoom(1234)
	}

	time.Sleep(time.Minute)
}
