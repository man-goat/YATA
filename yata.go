package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"nhooyr.io/websocket"
)

type post string

// what the fuck is a database
// yolo lmao
var posts []post

func main() {
	router := mux.NewRouter()

	posts = make([]post, 0)

	router.HandleFunc("/socket", handleSocket)
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	router.HandleFunc(
		"/posts",
		func(w http.ResponseWriter, r *http.Request) {
			text := strings.TrimSpace(r.FormValue("text"))
			if len(text) == 0 {
				return
			}
			p := post(text)
			SubscriptionHandler().Post(p)
			posts = append(posts, p)
		},
	).Methods("POST")
	router.HandleFunc("/", renderer)

	router.HandleFunc(
		"/posts",
		func(w http.ResponseWriter, r *http.Request) {
			// todo
		},
	).Methods("DELETE")

	http.ListenAndServe(":80", router)
}

func writeMsg(ctx context.Context, msg post, c *websocket.Conn) error {
	var rendered bytes.Buffer
	funcMap := template.FuncMap{
		"revIndex": func(index, length int) (revIndex int) { return (length - 1) - index },
	}
	tmpls, err := template.New("page.gohtml").Funcs(funcMap).ParseGlob("templates/*.gohtml")
	if err != nil {
		return err
	}
	if err := tmpls.ExecuteTemplate(&rendered, "response", msg); err != nil {
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
		log.Printf("unsub socket %s", sub.Key)
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

func renderer(w http.ResponseWriter, r *http.Request) {

	funcMap := template.FuncMap{
		"revIndex": func(index, length int) (revIndex int) { return (length - 1) - index },
	}

	tmpls, err := template.New("page.gohtml").Funcs(funcMap).ParseGlob("templates/*.gohtml")
	if err != nil {
		log.Print(err)
		return
	}

	// better way to do this?
	// https://gist.github.com/dmitshur/5f9e93c38f6b75421060
	err = tmpls.ExecuteTemplate(w, "page", posts)
	if err != nil {
		log.Print(err)
		return
	}
}
