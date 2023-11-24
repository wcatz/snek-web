package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/blinklabs-io/snek/event"
	filter_event "github.com/blinklabs-io/snek/filter/event"
	"github.com/blinklabs-io/snek/input/chainsync"
	output_embedded "github.com/blinklabs-io/snek/output/embedded"
	"github.com/blinklabs-io/snek/pipeline"
	"github.com/gorilla/websocket"
)

// HTML template
var templates *template.Template

// Initialize the HTML templates
func init() {
	templatesPath := filepath.Join(".", "templates", "*.html")
	templates = template.Must(template.ParseGlob(templatesPath))
}

// HTTP handler for rendering the HTML page
func handler(w http.ResponseWriter, r *http.Request) {
	// No need to create a BlockEvent here; it will be populated when a new event occurs.

	// Pass the TemplateData to the template
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// BlockEvent struct representing the structure of a block event
type BlockEvent struct {
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

// Define the WebSocket connection upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Maintain a map of connected WebSocket clients
var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

// Channel to broadcast block events to connected clients
var events = make(chan BlockEvent)

// Indexer struct to manage the Snek pipeline and block events
type Indexer struct {
	pipeline   *pipeline.Pipeline
	blockEvent BlockEvent
}

// Singleton instance of the Indexer
var globalIndexer = &Indexer{}

// WebSocket handler for broadcasting block events to connected clients
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Add the new client to the clients map
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	for {
		select {
		// Wait for a new block event to be sent to the events channel
		case blockEvent := <-events:
			// Serialize the block event to JSON
			message, err := json.Marshal(blockEvent)
			if err != nil {
				log.Println(err)
				continue
			}

			// Iterate over connected clients and send the message
			clientsMu.Lock()
			for client := range clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
			clientsMu.Unlock()
		}
	}
}

// Declare a temporary variable to capture the result of the function call
var node = chainsync.WithAddress("backbone.cardano-mainnet.iohk.io:3001")

// Options for the ChainSync input
var inputOpts = []chainsync.ChainSyncOptionFunc{
	node,
	chainsync.WithNetworkMagic(764824073),
	chainsync.WithIntersectTip(true),
}
// Start the Snek pipeline and handle block events
func (i *Indexer) Start() error {
	// Create a new pipeline
	i.pipeline = pipeline.New()

	// Configure ChainSync input
	input_chainsync := chainsync.New(inputOpts...)
	i.pipeline.AddInput(input_chainsync)

	// Configure filter to handle only block events
	filterEvent := filter_event.New(filter_event.WithTypes([]string{"chainsync.block"}))
	i.pipeline.AddFilter(filterEvent)

	// Configure embedded output with callback function
	output := output_embedded.New(output_embedded.WithCallbackFunc(i.handleEvent))
	i.pipeline.AddOutput(output)

	// Start the pipeline
	if err := i.pipeline.Start(); err != nil {
		log.Fatalf("failed to start pipeline: %s\n", err)
	}

	// Start error handler in a goroutine
	go func() {
		err, ok := <-i.pipeline.ErrorChan()
		if ok {
			log.Fatalf("pipeline failed: %s\n", err)
		}
	}()

	return nil
}

// Handle block events received from the Snek pipeline
func (i *Indexer) handleEvent(event event.Event) error {
	// Marshal the event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Unmarshal JSON data into BlockEvent struct
	var blockEvent BlockEvent
	err = json.Unmarshal(data, &blockEvent)
	if err != nil {
		return err
	}

	// Format the timestamp into a human-readable form
	parsedTime, err := time.Parse(time.RFC3339, blockEvent.Timestamp)
	if err == nil {
		blockEvent.Timestamp = parsedTime.Format("January 2, 2006 15:04:05 MST")
	}

	// Update the blockEvent field in the Indexer
	i.blockEvent = blockEvent

	// Print the block event struct to the console
	fmt.Printf("Received BlockEvent: %+v\n", blockEvent)

	// Send the block event to the WebSocket clients
	events <- blockEvent

	return nil
}

// Main function to start the Snek pipeline and serve HTTP requests
func main() {
	// Start the Snek pipeline
	if err := globalIndexer.Start(); err != nil {
		log.Fatalf("failed to start snek: %s", err)
	}

	// Define HTTP handlers

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)

	// Start the HTTP server on port 8080
	http.ListenAndServe(":8080", nil)
}
