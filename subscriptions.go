package main

import (
	"log"

	"github.com/google/uuid"
)

type SubKey string
type Subscription struct {
	Key      SubKey
	Listener chan PostCrudEvent
}

type subscriptionHandler struct {
	sub   chan Subscription
	unsub chan SubKey

	notify chan PostCrudEvent
	subs   map[SubKey]chan PostCrudEvent
}

var handler = subscriptionHandler{
	sub:    make(chan Subscription),
	unsub:  make(chan SubKey),
	notify: make(chan PostCrudEvent),
	subs:   make(map[SubKey]chan PostCrudEvent),
}

func GenerateKey() SubKey {
	return SubKey(uuid.New().String())
}

func Subscribe(listener chan PostCrudEvent) SubKey {
	sub := Subscription{Listener: listener, Key: GenerateKey()}
	handler.sub <- sub
	return sub.Key
}

func Unsubscribe(sub SubKey) {
	handler.unsub <- sub
}

func BroadcastEvent(event PostCrudEvent) {
	log.Printf("broadcast: %s", event)
	handler.notify <- event
}

func init() {
	go func() {
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
	}()
}
