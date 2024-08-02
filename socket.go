package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

func writePost(ctx context.Context, c *websocket.Conn, post Post) error {
	cx2, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	writer, err := c.Writer(cx2, websocket.MessageText)
	if err != nil {
		return err
	}
	defer writer.Close()
	return writeTemplate(writer, post, "new")
}

// is this a good idea?
func deletePost(ctx context.Context, c *websocket.Conn, post Post) error {
	cx2, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	writer, err := c.Writer(cx2, websocket.MessageText)
	if err != nil {
		return err
	}
	defer writer.Close()
	return writeTemplate(writer, post, "delete")
}

func handleSocket(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	defer c.CloseNow()

	ctx := r.Context()
	ctx = c.CloseRead(ctx)

	listener := make(chan PostCrudEvent, 100)
	key := Subscribe(listener)

	defer func() {
		log.Printf("unsub socket %s", key)
		Unsubscribe(key)
	}()

	for {
		select {
		case msg := <-listener:
			switch msg.Type {
			case CreateEvent:
				err = writePost(ctx, c, msg.Payload)
				if err != nil {
					fmt.Println(err)
					return
				}
			case DeleteEvent:
				err = deletePost(ctx, c, msg.Payload)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
