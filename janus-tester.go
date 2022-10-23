package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Hwanse/janus-tester/internal"
	"github.com/Hwanse/janus-tester/internal/janus"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {

	fileFlag := flag.String("f", "test-sample.json", "input test scenario sample ")
	flag.Parse()

	fmt.Println("read sample file : ", *fileFlag)

	data, err := os.ReadFile(*fileFlag)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	scenario := Scenario{}
	err = json.Unmarshal(data, &scenario)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("%+v \n", scenario)

	url := fmt.Sprintf("ws://%s:%s/", janus.JanusLocalHost, janus.JanusWebsocketPort)
	gateway, err := janus.WsConnect(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	session, err := gateway.Create()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	go func(ctx context.Context, session *janus.Session) {
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
	}(context.Background(), session)

	handle, err := session.Attach(janus.VideoRoomPluginName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ctx, destroy := context.WithCancel(context.Background())
	roomList := make([]uint64, 0)
	wg := &sync.WaitGroup{}
	endSignal := make(chan os.Signal, 1)
	signal.Notify(endSignal, os.Interrupt)

	for _, roomScenario := range scenario.RoomScenarios {
		roomID, err := CreateRoom(handle, roomScenario.PublisherLimitCount)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		roomList = append(roomList, roomID)

		for i := 0; i < roomScenario.ActivePublisherCount; i++ {
			wg.Add(1)
			go AttachPublisher(ctx, gateway, roomID, wg)
		}

		for i := 0; i < roomScenario.SubscriberCount; i++ {
			wg.Add(1)
			go AttachSubscriber(ctx, gateway, roomID, wg)
		}
	}

	<-endSignal
	log.Println("process end signal, destroy all peers & rooms")

	destroy()
	wg.Wait()

	for _, id := range roomList {
		RemoveRoom(handle, id)
	}
}

type Scenario struct {
	Description   string
	RoomScenarios []RoomScenario `json:"room_scenarios"`
}

type RoomScenario struct {
	PublisherLimitCount  int `json:"publisher_limit_count"`
	ActivePublisherCount int `json:"active_publisher_count"`
	SubscriberCount      int `json:"subscriber_count"`
	JoinTimeInterval     int `json:"join_time_interval"`
}

func CreateRoom(handle *janus.Handle, publisherLimitCount int) (uint64, error) {
	rand.Seed(time.Now().UnixNano())

	roomID := uint64(rand.Uint32())
	req := &janus.CreateRoomRequest{
		Request: janus.TypeCreate,
		Room: janus.Room{
			RoomID:              roomID,
			IsPrivate:           false,
			PublisherLimitCount: publisherLimitCount,
			UseRecord:           false,
			NotifyJoining:       false,
			Bitrate:             128000,
			BitrateCap:          false,
		},
	}

	err := handle.CreateRoom(req)
	if err != nil {
		return 0, err
	}

	return roomID, nil
}

func RemoveRoom(handle *janus.Handle, roomID uint64) error {
	req := &janus.DestroyRoomRequest{
		Request: janus.TypeDestroy,
		RoomID:  roomID,
	}

	return handle.DestroyRoom(req)
}

func AttachSubscriber(ctx context.Context, gateway *janus.Gateway, roomID uint64, wg *sync.WaitGroup) {
	session, err := gateway.Create()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	client := internal.NewClient(session)

	client.JoinRoom(ctx, roomID)
	go client.KeepAliveLoop(ctx)

	client.KeepConnection(ctx)
	defer func() {
		client.LeaveRoom()
		wg.Done()
	}()
}

func AttachPublisher(ctx context.Context, gateway *janus.Gateway, roomID uint64, wg *sync.WaitGroup) {
	session, err := gateway.Create()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	client := internal.NewClient(session)

	client.JoinRoom(ctx, roomID)
	go client.KeepAliveLoop(ctx)

	client.TestPublishStream(ctx)

	client.KeepConnection(ctx)
	defer func() {
		client.LeaveRoom()
		wg.Done()
	}()
}
