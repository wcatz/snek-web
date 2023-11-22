// main.go

package main

import (
	"html/template"
	"log"
	"net/http"

	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
//	events    = make(chan SnekEvent)
	templates *template.Template
)

func init() {
	templatesPath := filepath.Join(".", "templates", "*.html")
	templates = template.Must(template.ParseGlob(templatesPath))
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", r.Host)
	if err != nil {
		log.Fatal(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

}

func main() {

	// Start the WebSocket server
	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
