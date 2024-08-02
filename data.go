package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type PostKey string
type Post struct {
	Content string
	Id      PostKey
}

type CrudEventType int8

const (
	CreateEvent CrudEventType = iota
	DeleteEvent
	UpdateEvent
)

type PostCrudEvent struct {
	Type    CrudEventType
	Payload Post
}

func CreatePost(content string) Post {
	post := Post{Content: content, Id: PostKey(uuid.New().String())}

	postMutex.Lock()
	posts = append(posts, post)
	postMutex.Unlock()
	return post
}

func Posts() []Post {
	return posts
}

var postMutex = new(sync.Mutex)
var posts = make([]Post, 0)

func (p Post) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\"%s\"", string(p.Id)))
	b.WriteString(" ")
	b.WriteString(fmt.Sprintf("%.10s", p.Content))
	return b.String()
}

func (e PostCrudEvent) String() string {
	var b strings.Builder

	var eventType string = "invalid type"
	switch e.Type {
	case CreateEvent:
		eventType = "CREATE"
	case DeleteEvent:
		eventType = "DELETE"
	case UpdateEvent:
		eventType = "UPDATE"
	}
	b.WriteString(
		fmt.Sprintf("%s: %s", eventType, e.Payload),
	)
	return b.String()
}
