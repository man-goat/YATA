package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"nhooyr.io/websocket"
)

type post string

func main() {
	// what the fuck is a database
	// yolo lmao
	posts := make([]post, 0)
	mux := http.NewServeMux()

	mux.HandleFunc("/socket", handleSocket)

	mux.HandleFunc("POST /add", func(w http.ResponseWriter, r *http.Request) {
		p := post(r.FormValue("text"))
		tmpl := template.Must(template.ParseFiles("templates/page.html"))
		SubscriptionHandler().Notify(p)
		posts = append(posts, p)
		err := tmpl.Execute(w, posts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/page.html"))
		err := tmpl.Execute(w, posts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	})
	http.ListenAndServe(":80", mux)
}

func writeMsg(ctx context.Context, msg post, c *websocket.Conn) error {
	var rendered bytes.Buffer
	mtmpl := template.Must(template.ParseFiles("templates/post.html"))
	if err := mtmpl.Execute(&rendered, msg); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, rendered.Bytes())
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

	sessKey := SubscriptionHandler().GenerateKey()
	sub := subscription{Key: sessKey, Listener: make(chan post, 1)}

	SubscriptionHandler().Subscribe(sub)

	defer func() {
		fmt.Printf("unsub socket %s\n", sub.Key)
		SubscriptionHandler().Unsubscribe(sub.Key)
	}()

	for {
		select {
		case msg := <-sub.Listener:
			err = writeMsg(ctx, msg, c)
			if err != nil {
				fmt.Println(err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
