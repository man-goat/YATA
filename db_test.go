package main

import (
	"context"
	"testing"
)

func TestDB(t *testing.T) {
	err := InitDB()
	if err != nil {
		t.Fatal(err)
	}

	postContent := "blarg"
	post, err := CreatePost(context.Background(), postContent)
	if err != nil {
		t.Fatal(err)
	}

	posts, err := GetPosts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatal("no rows returned")
	}
	if posts[0].Content != postContent {
		t.Fatal("Did not find key in db table")
	}

	_, err = DeletePost(context.Background(), post)
	if err != nil {
		t.Fatal(err)
	}
	posts, err = GetPosts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatal("not removed")
	}
	res, err := DeletePost(context.Background(), post)
	if err != nil {
		t.Fatal(err)
	}
	n, err := res.RowsAffected()
	if n != 0 {
		t.Fatal("second delete should affect 0 rows")
	}
}
