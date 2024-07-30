package main

import (
	"log"
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

func (handler *subscriptionHandler) Post(p post) {
	handler.notify <- p
}

func (handler *subscriptionHandler) handleSubs() {
	for {
		select {
		case ns := <-handler.sub:
			log.Print("new sub")
			handler.subs[ns.Key] = ns.Listener
		case msg := <-handler.notify:
			log.Print("new message, notifying")
			for key, sub := range handler.subs {
				log.Printf("notifying %s", key)
				select {
				case sub <- msg:
				default:
					log.Print("... no response")
				}
			}
		case unsubKey := <-handler.unsub:
			log.Printf("unsub: %s", unsubKey)
			delete(handler.subs, unsubKey)
		}
	}
}
