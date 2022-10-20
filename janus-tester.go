package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Hwanse/janus-tester/internal"
	"github.com/Hwanse/janus-tester/internal/janus"
	"os"
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

	duration := time.Duration(scenario.Duration)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*duration)

	wg := &sync.WaitGroup{}
	for i := 0; i < scenario.SubscriberCount; i++ {
		wg.Add(1)
		go AttachSubscriber(ctx, wg)
	}

	wg.Wait()
}

type Scenario struct {
	SubscriberCount int `json:"subscriber_count"`
	Duration        int
	Description     string
}

func AttachSubscriber(ctx context.Context, wg *sync.WaitGroup) {
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

	client := internal.NewClient(session)

	client.JoinRoom(ctx, 1234)
	go client.KeepAliveLoop(ctx)

	client.KeepConnection(ctx)
	wg.Done()
}
