package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

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
			p := CreatePost(text)

			BroadcastEvent(PostCrudEvent{Payload: p, Type: CreateEvent})
		},
	).Methods("POST")
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := writeTemplate(w, Posts(), "page")
		if err != nil {
			log.Print(err)
		}
	})

	router.HandleFunc(
		"/posts",
		func(w http.ResponseWriter, r *http.Request) {
			postId := r.URL.Query().Get("id")

			payload := PostCrudEvent{Type: DeleteEvent, Payload: Post{Id: PostKey(postId)}}
			BroadcastEvent(payload)

			w.WriteHeader(204)
			// todo think about concurrent delete problem
			// todo actually delete lmao
		},
	).Methods("DELETE")

	http.ListenAndServe(":80", router)
}
