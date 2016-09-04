package sse

import (
	"github.com/mijia/sweb/log"

	"encoding/json"
	"fmt"
	"net/http"
)

type Event struct {
	Id    int64       `json:"id"`
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func (e *Event) String() string {
	if data, err := json.Marshal(e.Data); err == nil {
		return fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", e.Id, e.Event, string(data))
	}
	return fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", e.Id, e.Event, e.Data)
}

type conn struct {
	w      http.ResponseWriter
	closed chan struct{}
}

func NewConn(w http.ResponseWriter) *conn {
	return &conn{w: w, closed: make(chan struct{}, 1)}
}

type SseHandler struct {
	id      int64
	event   chan interface{}
	clients map[*conn]struct{}
}

func NewSseHandler(event chan interface{}) *SseHandler {
	return &SseHandler{
		id:      int64(0),
		event:   event,
		clients: make(map[*conn]struct{}),
	}
}

func (sse *SseHandler) StartSendEvent() {
	for {
		data := <-sse.event
		sse.sendEvent(data)
	}
}

func (sse *SseHandler) sendEvent(data interface{}) {
	sse.id++
	for conn, _ := range sse.clients {
		e := &Event{Id: sse.id, Event: "update", Data: data}
		flush := conn.w.(http.Flusher)
		if _, err := fmt.Fprintf(conn.w, e.String()); err != nil {
			log.Error("write err:", err)
			conn.closed <- struct{}{}
		} else {
			flush.Flush()
		}
	}
}

func MakeHeaderAsSSEWriter(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func (sse *SseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	MakeHeaderAsSSEWriter(w)
	c := NewConn(w)
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		c.closed <- struct{}{}
	}()
	sse.commingConnection(c)
	sse.closeConnection(c)
}

func (sse *SseHandler) commingConnection(c *conn) {
	sse.clients[c] = struct{}{}
}

func (sse *SseHandler) closeConnection(c *conn) {
	defer close(c.closed)
	<-c.closed
	log.Info("closeConnection")
	delete(sse.clients, c)
}
