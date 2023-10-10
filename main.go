package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var addr = flag.String("addr", "127.0.0.1", "http service address")
var port = flag.Int("port", 6611, "http server port")
var staticDir = flag.String("static", "static", "static files directory")
var fileServer http.Handler
var games = make(map[string]*Game)

func serveGame(w http.ResponseWriter, r *http.Request) {
	log.Println("Connection to", r.URL.Path)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameid := strings.TrimPrefix(r.URL.Path, "/game/")

	if len(gameid) == 0 {
		game := newGame(games)
		http.Redirect(w, r, r.URL.Path+game.id, http.StatusSeeOther)
		return
	}

	if _, ok := games[gameid]; ok {
		http.ServeFile(w, r, filepath.Join(*staticDir, "server/index.html"))
		return
	} else {
		http.Redirect(w, r, "/404/", http.StatusSeeOther)
		return
	}
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	log.Println("Connection to", r.URL.Path)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if fileServer == nil {
		log.Panicln("no file server")
	}

	if _, err := os.Stat(filepath.Join(*staticDir, r.URL.Path)); err != nil {
		http.Redirect(w, r, "/404/", http.StatusSeeOther)
		return
	}

	fileServer.ServeHTTP(w, r)
}

func serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("Connection to", r.URL.Path)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameid := strings.TrimPrefix(r.URL.Path, "/ws/")

	if game, ok := games[gameid]; ok {
		game.serveWS(w, r)
		return
	} else {
		http.Redirect(w, r, "/404/", http.StatusSeeOther)
		return
	}
}

func serveNotFound(w http.ResponseWriter, r *http.Request) {
	log.Println("Connection to", r.URL.Path)
	http.ServeFile(w, r, filepath.Join(*staticDir, "404.html"))
}

func oldServer() {
	flag.Parse()

	fileServer = http.FileServer(http.Dir(*staticDir))

	http.HandleFunc("/game/", serveGame)
	http.HandleFunc("/ws/", serveWS)
	http.HandleFunc("/404/", serveNotFound)
	http.HandleFunc("/", serveFiles)

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", *addr, *port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Printf("Server listening on http://%s", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	oldServer()
}
