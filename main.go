package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var hubs = make(map[string]*Hub)

// func serveHome(w http.ResponseWriter, r *http.Request) {
// 	log.Println("Connection to home")
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	http.Redirect(w, r, r.URL.Path+"/room/", http.StatusSeeOther)
// }

func serveRoom(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/room/")

	if len(id) == 0 {
		newRoom(w, r)
		return
	}

	log.Println("Connection to room", id)

	if _, ok := hubs[id]; ok {
		http.ServeFile(w, r, "home.html")
		return
	} else {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
}

func newRoom(w http.ResponseWriter, r *http.Request) {
	hub := newHub()
	hubs[hub.id] = hub
	go hub.run(hubs)
	http.Redirect(w, r, r.URL.Path+hub.id, http.StatusSeeOther)
}

func main() {
	flag.Parse()

	http.HandleFunc("/room/", serveRoom)

	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/ws/")
		serveWS(hubs[id], w, r)
	})

	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Println("Starting server")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
