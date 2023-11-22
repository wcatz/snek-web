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
	templates *template.Template
)

type pipeline struct {
	input  chan interface{}
	output chan interface{}
}

func newPipeline() *pipeline {
	return &pipeline{
		input:  make(chan interface{}),
		output: make(chan interface{}),
	}
}

func (p *pipeline) send(data interface{}) {
	p.input <- data
}

func (p *pipeline) receive() <-chan interface{} {
	return p.output
}

func (p *pipeline) start() {
	go func() {
		for {
			data := <-p.input
			// You can perform any processing on 'data' here
			// For simplicity, just forward it to the output channel
			p.output <- data

			// Send the data to all connected clients
			clientsMu.Lock()
			for client := range clients {
				err := client.WriteMessage(websocket.TextMessage, []byte(data.(string)))
				if err != nil {
					log.Printf("Error sending message to client: %v", err)
				}
			}
			clientsMu.Unlock()
		}
	}()
}

func init() {
	templatesPath := filepath.Join(".", "templates", "*.html")
	templates = template.Must(template.ParseGlob(templatesPath))
}

func renderError(w http.ResponseWriter, err error, statusCode int) {
	log.Printf("Error: %v", err)
	http.Error(w, http.StatusText(statusCode), statusCode)
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", r.Host)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}
}

var dataPipeline *pipeline

func init() {
	dataPipeline = newPipeline()
	dataPipeline.start()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing connection: %v", closeErr)
		}
	}()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	// Handle WebSocket messages here
	go handleWebSocketMessages(conn)
}

func handleWebSocketMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		// Send the received message through the pipeline
		dataPipeline.send(string(message))
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)

	addr := ":8080"
	log.Printf("Server started on %s\n", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
