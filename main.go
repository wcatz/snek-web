// main.go

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	events    = make(chan SnekEvent)
	templates *template.Template
)

type SnekEvent struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Context   struct {
		BlockNumber  int `json:"blockNumber"`
		SlotNumber   int `json:"slotNumber"`
		NetworkMagic int `json:"networkMagic"`
	} `json:"context"`
	Payload struct {
		BlockBodySize    int    `json:"blockBodySize"`
		IssuerVkey       string `json:"issuerVkey"`
		BlockHash        string `json:"blockHash"`
		TransactionCount int    `json:"transactionCount"`
	} `json:"payload"`
}

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

	for {
		select {
		case snekEvent := <-events:
			clientsMu.Lock()
			for client := range clients {
				// Check if the WebSocket connection is still open
				if err := client.WriteJSON(snekEvent); err != nil {
					if websocket.IsCloseError(err, websocket.CloseNormalClosure) || websocket.IsCloseError(err, websocket.CloseGoingAway) {
						// Connection is closed by the client, remove it from the clients map
						delete(clients, client)
					} else {
						log.Println("Error writing to client:", err)
					}
				}
			}
			clientsMu.Unlock()
		}
	}
}

func startSnek() {
	cmd := exec.Command("snek",
		"-input-chainsync-address", "m2:6002",
		"-filter-type", "chainsync.block",
		"-output", "webhook",
		"-output-webhook-url", "http://localhost:42069/webhook")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("Error creating stdout pipe:", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal("Error starting snek:", err)
		return
	}

	fmt.Println("Snek started successfully. PID:", cmd.Process.Pid)

	decoder := json.NewDecoder(stdout)
	for {
		var snekEvent SnekEvent
		if err := decoder.Decode(&snekEvent); err != nil {
			log.Println("Error decoding snek output:", err)
			break
		}
		events <- snekEvent
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var snekEvent SnekEvent
	if err := json.NewDecoder(r.Body).Decode(&snekEvent); err != nil {
		log.Println("Error decoding webhook data:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Do something with the webhook data if needed
	fmt.Printf("Received webhook data: %+v\n", snekEvent)

	// Send the event to WebSocket clients
	clientsMu.Lock()
	for client := range clients {
		err := client.WriteJSON(snekEvent)
		if err != nil {
			log.Println("Error writing to client:", err)
		}
	}
	clientsMu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func main() {
	go startSnek()

	// Set up the Gorilla Mux router for the webhook server on port 42069
	webhookRouter := mux.NewRouter()
	webhookRouter.HandleFunc("/webhook", handleWebhook).Methods(http.MethodPost)

	// Start the webhook HTTP server on port 42069
	webhookPort := 42069
	webhookAddr := fmt.Sprintf(":%v", webhookPort)
	fmt.Printf("Webhook server running on port %v...\n", webhookPort)

	go func() {
		err := http.ListenAndServe(webhookAddr, webhookRouter)
		if err != nil {
			fmt.Println("Error starting webhook HTTP server:", err)
		}
	}()

	// Start the WebSocket server
	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
