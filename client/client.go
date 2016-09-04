package main

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/supermeng/gosse/sse"
	"golang.org/x/net/context"
)

type Client struct {
	Addr string
}

func NewClient(Addr string) *Client {
	return &Client{Addr: Addr}
}

func (c *Client) Do(uri string) (io.ReadCloser, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	u.Scheme = "http"
	u.Host = c.Addr

	resp, err := (&http.Client{}).Get(u.String())
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *Client) Watch(uri string, ctx context.Context) (<-chan *sse.Event, error) {
	reader, err := c.Do(uri)
	if err != nil {
		return nil, err
	}
	bufReader := bufio.NewReader(reader)
	respCh := make(chan *sse.Event, 1)
	stop := make(chan struct{}, 1)
	go func() {
		select {
		case <-ctx.Done():
		case <-stop:
		}
		reader.Close()
	}()
	go func() {
		defer close(stop)
		defer close(respCh)
		newEvent := true
		var e *sse.Event
		for {
			if newEvent {
				e = new(sse.Event)
				newEvent = false
			}
			line, err := bufReader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF || len(line) == 0 {
					return
				}
			}
			if pureLine := bytes.TrimSpace(line); len(pureLine) == 0 {
				newEvent = true
				respCh <- e
			} else {
				fields := bytes.SplitN(pureLine, []byte{':'}, 2)
				if len(fields) < 2 {
					continue
				}
				switch key := string(bytes.TrimSpace(fields[0])); key {
				case "id":
					if id, err := strconv.ParseInt(string(bytes.TrimSpace(fields[1])), 10, 64); err == nil {
						e.Id = id
					}
				case "event":
					e.Event = string(bytes.TrimSpace(fields[1]))
				case "data":
					e.Data = bytes.TrimSpace(fields[1])
				}
			}
		}
	}()
	return respCh, nil
}
