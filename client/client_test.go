package main

import (
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"

	"testing"
)

func Test_Watch(t *testing.T) {
	c := NewClient("127.0.0.1:8888")
	ctx := context.Background()
	if respch, err := c.Watch("/object", ctx); err != nil {
		log.Error(err)
	} else {
		for event := range respch {
			log.Info(event)
		}
	}
}
