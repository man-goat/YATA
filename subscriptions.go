package main

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

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

func init() {
	SubscriptionHandler()
}

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
