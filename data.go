package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
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

var db *sql.DB

func init() {
	err := InitDB()
	if err != nil {
		panic(err)
	}
}

func InitDB() error {
	var err error
	// you can have a persistent store when you're old enough
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		return err
	}
	_, err = db.ExecContext(context.Background(),
		`create table messages (id uuid primary key, content text not null)`,
	)
	if err != nil {
		return err
	}

	return nil
}

func CreatePost(ctx context.Context, content string) (Post, error) {
	post := Post{Content: content, Id: PostKey(uuid.New().String())}

	_, err := db.ExecContext(ctx, "insert into messages (id, content) values (?, ?)", post.Id, post.Content)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func DeletePost(ctx context.Context, remove Post) (sql.Result, error) {
	return db.ExecContext(ctx, "delete from messages where id = ?", remove.Id)
}

func GetPosts(ctx context.Context) ([]Post, error) {
	rows, err := db.Query(`select id, content from messages`)
	if err != nil {
		return nil, err
	}
	var out []Post

	for rows.Next() {
		var post Post
		if err = rows.Scan(&post.Id, &post.Content); err != nil {
			return nil, err
		}
		out = append(out, post)
	}

	return out, nil
}

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
