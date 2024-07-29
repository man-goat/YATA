package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type post string
type subkey string
type subscription struct {
	Key      subkey
	Listener chan post
}

type subscriptionHandler struct {
	sub   chan subscription
	unsub chan subkey

	notify chan post
	subs   map[subkey]chan post
}

var handler *subscriptionHandler = nil
var handlerLock = &sync.Once{}

func SubscriptionHandler() *subscriptionHandler {
	if handler != nil {
		return handler
	}
	handlerLock.Do(
		func() {
			handler = &subscriptionHandler{
				sub:    make(chan subscription),
				unsub:  make(chan subkey),
				notify: make(chan post),
				subs:   make(map[subkey]chan post),
			}
			go handler.handleSubs()
		},
	)
	return handler
}

func (*subscriptionHandler) GenerateKey() subkey {
	return subkey(uuid.New().String())
}

func (handler *subscriptionHandler) Subscribe(sub subscription) {
	handler.sub <- sub
}

func (handler *subscriptionHandler) Unsubscribe(sub subkey) {
	handler.unsub <- sub
}

func (handler *subscriptionHandler) Notify(p post) {
	handler.notify <- p
}

func (handler *subscriptionHandler) handleSubs() {
	for {
		select {
		case ns := <-handler.sub:
			fmt.Println("new sub")
			handler.subs[ns.Key] = ns.Listener
		case msg := <-handler.notify:
			fmt.Println("new message, notifying")
			for key, sub := range handler.subs {
				fmt.Printf("notifying %s", key)
				select {
				case sub <- msg:
				default:
					fmt.Print("... no response")
				}
				fmt.Println()
			}
		case unsubKey := <-handler.unsub:
			fmt.Printf("unsub: %s\n", unsubKey)
			delete(handler.subs, unsubKey)
		}
	}
}

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
	SubscriptionHandler()
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
