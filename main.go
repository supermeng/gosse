package main

import (
	"github.com/mijia/sweb/log"
	"github.com/supermeng/gosse/sse"
	"golang.org/x/net/netutil"

	"net"
	"net/http"
	"time"
)

const (
	MAX_CONN = 1000
)

type TestObj struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	age  int    `json:"age"`
}

func HelloWroldEvent(event chan<- interface{}) {
	for {
		time.Sleep(2 * time.Second)
		event <- "Hello world!"
	}
}

func SlicesEvent(event chan<- interface{}) {
	for {
		time.Sleep(2 * time.Second)
		a := []int{1, 2, 3}
		event <- a
	}
}

func ObjectEvent(event chan<- interface{}) {
	for {
		time.Sleep(2 * time.Second)
		a := TestObj{Id: 1, Name: "test"}
		event <- a
	}

}

type GenEvent func(event chan<- interface{})

func startSSE(event chan interface{}, genEvent GenEvent) *sse.SseHandler {
	sseHandler := sse.NewSseHandler(event)

	go sseHandler.StartSendEvent()
	go genEvent(event)
	return sseHandler
}

func main() {
	hwEvent := make(chan interface{}, 1)
	SlEvent := make(chan interface{}, 1)
	obEvent := make(chan interface{}, 1)

	helloHandler := startSSE(hwEvent, HelloWroldEvent)
	sliceHandler := startSSE(SlEvent, SlicesEvent)
	objectHandler := startSSE(obEvent, ObjectEvent)

	http.Handle("/helloworld", helloHandler)
	http.Handle("/slice", sliceHandler)
	http.Handle("/object", objectHandler)

	l, err := net.Listen("tcp", "0.0.0.0:8888")
	if err != nil {
		log.Fatal("Listen: %v\n", err)
	}
	defer l.Close()
	l = netutil.LimitListener(l, MAX_CONN)

	log.Fatal(http.Serve(l, nil))
}
