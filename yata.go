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
			p, err := CreatePost(r.Context(), text)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Print(err)
			}

			BroadcastEvent(PostCrudEvent{Payload: p, Type: CreateEvent})
		},
	).Methods("POST")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := GetPosts(r.Context())
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = writeTemplate(w, posts, "page")
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	router.HandleFunc(
		"/posts",
		func(w http.ResponseWriter, r *http.Request) {
			postId := r.URL.Query().Get("id")

			_, err := DeletePost(r.Context(), Post{Id: PostKey(postId)})
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			payload := PostCrudEvent{Type: DeleteEvent, Payload: Post{Id: PostKey(postId)}}
			BroadcastEvent(payload)

			w.WriteHeader(204)
			// todo think about concurrent delete problem
			// todo actually delete lmao
		},
	).Methods("DELETE")

	http.ListenAndServe(":80", router)
}
