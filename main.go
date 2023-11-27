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

// Mutex to synchronize access to the node address
var nodeMu sync.Mutex

// Node address as a string
//var nodeAddress string

// Initialize the HTML templates
func init() {
	templatesPath := filepath.Join(".", "templates", "*.html")
	templates = template.Must(template.ParseGlob(templatesPath))
}

// TemplateData holds data for the HTML template
type TemplateData struct {
	NodeAddress string
}

type BlockEvent struct {
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Context   chainsync.BlockContext `json:"context"`
	Payload   chainsync.BlockEvent   `json:"payload"`
}

type RollbackEvent struct {
	Type      string                  `json:"type"`
	Timestamp string                  `json:"timestamp"`
	Payload   chainsync.RollbackEvent `json:"payload"`
}

type TransactionEvent struct {
	Type      string                       `json:"type"`
	Timestamp string                       `json:"timestamp"`
	Context   chainsync.TransactionContext `json:"context"`
	Payload   chainsync.TransactionEvent   `json:"payload"`
}

// HTTP handler for rendering the HTML page
func handler(w http.ResponseWriter, r *http.Request) {
	// Create an instance of TemplateData with the current node address
	node := TemplateData{
		NodeAddress: globalIndexer.nodeAddress,
	}

	// Pass the TemplateData to the template
	if err := templates.ExecuteTemplate(w, "index.html", node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// type BlockEvent struct {
// 	Type      string                 `json:"type"`
// 	Timestamp string                 `json:"timestamp"`
// 	Context   chainsync.BlockContext `json:"context"`
// 	Payload   chainsync.BlockEvent   `json:"payload"`
// }

// Define the WebSocket connection upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Maintain a map of connected WebSocket clients
var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

// Channel to broadcast block events to connected clients
// var events = make(chan BlockEvent)
var events = make(chan interface{})

// Indexer struct to manage the Snek pipeline and block events
type Indexer struct {
	pipeline         *pipeline.Pipeline
	blockEvent       BlockEvent
	rollbackEvent    RollbackEvent
	transactionEvent TransactionEvent
	nodeAddress      string
	eventType        string
	isRunning        bool
}

// Singleton instance of the Indexer
var globalIndexer = &Indexer{
	nodeAddress: "backbone.cardano-mainnet.iohk.io:3001", // Default address

}

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
		case event := <-events:
			// Serialize the block event to JSON
			message, err := json.Marshal(event)
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

// Start the Snek pipeline and handle block events
func (i *Indexer) Start() error {
	// Define node and inputOpts inside Start, using the current node address
	node := chainsync.WithAddress(i.nodeAddress)
	inputOpts := []chainsync.ChainSyncOptionFunc{
		node,
		chainsync.WithNetworkMagic(764824073),
		chainsync.WithIntersectTip(true),
	}
	// Create a new pipeline
	i.pipeline = pipeline.New()

	// Configure ChainSync input
	input_chainsync := chainsync.New(inputOpts...)
	i.pipeline.AddInput(input_chainsync)

	// Configure filter to handle only block events
	// Update the event type filter based on the selection
	filterEvent := filter_event.New(filter_event.WithTypes([]string{i.eventType}))
	i.pipeline.AddFilter(filterEvent)

	// Configure zembedded output with callback function
	output := output_embedded.New(output_embedded.WithCallbackFunc(i.handleEvent))
	i.pipeline.AddOutput(output)

	// Start the pipeline
	if err := i.pipeline.Start(); err != nil {
		log.Printf("failed to start pipeline: %s\n", err)
		return fmt.Errorf("failed to start pipeline: %w", err)
	}

	// Start error handler in a goroutine
	go func() {
		err, ok := <-i.pipeline.ErrorChan()
		if ok {
			log.Printf("pipeline failed: %s\n", err)
		}
	}()

	i.isRunning = true

	return nil
}

// Handle block events received from the Snek pipeline
func (i *Indexer) handleEvent(event event.Event) error {

	// Marshal the event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	var getEvent map[string]interface{}
	errr := json.Unmarshal(data, &getEvent)
	if errr != nil {
		return err
	}

	eventType, ok := getEvent["type"].(string)
	if !ok {
		return fmt.Errorf("failed to get event type")
	}

	switch eventType {
	case "chainsync.block":
		var blockEvent BlockEvent
		err := json.Unmarshal(data, &blockEvent)
		if err != nil {
			return err
		}

		// Format the timestamp into a human-readable form
		parsedTime, err := time.Parse(time.RFC3339, blockEvent.Timestamp)
		if err == nil {
			blockEvent.Timestamp = parsedTime.Format("January 2, 2006 15:04:05 MST")
		}

		// Update the currentEvent field in the Indexer
		i.blockEvent = blockEvent

		// Print the block event struct to the console
		fmt.Printf("Received Event: %+v\n", blockEvent)

		// Send the block event to the WebSocket clients
		events <- blockEvent
	case "chainsync.rollback":
		var rollbackEvent RollbackEvent
		err := json.Unmarshal(data, &rollbackEvent)
		if err != nil {
			return err
		}

		// Format the timestamp into a human-readable form
		parsedTime, err := time.Parse(time.RFC3339, rollbackEvent.Timestamp)
		if err == nil {
			rollbackEvent.Timestamp = parsedTime.Format("January 2, 2006 15:04:05 MST")
		}

		// Update the currentEvent field in the Indexer
		i.rollbackEvent = rollbackEvent

		// Print the rollbackk event struct to the console
		fmt.Printf("Received Event: %+v\n", rollbackEvent)

		// Send the block event to the WebSocket clients
		events <- rollbackEvent
	case "chainsync.transaction":
		var transactionEvent TransactionEvent
		err := json.Unmarshal(data, &transactionEvent)
		if err != nil {
			return err
		}

		// Format the timestamp into a human-readable form
		parsedTime, err := time.Parse(time.RFC3339, transactionEvent.Timestamp)
		if err == nil {
			transactionEvent.Timestamp = parsedTime.Format("January 2, 2006 15:04:05 MST")
		}

		// Update the currentEvent field in the Indexer
		i.transactionEvent = transactionEvent

		// Print the transaction event struct to the console
		fmt.Printf("Received Event: %+v\n", transactionEvent)

		// Send the block event to the WebSocket clients
		events <- transactionEvent
	}

	// // Unmarshal JSON data into BlockEvent struct
	// var blockEvent BlockEvent
	// err = json.Unmarshal(data, &blockEvent)
	// if err != nil {
	// 	return err
	// }

	// // Format the timestamp into a human-readable form
	// parsedTime, err := time.Parse(time.RFC3339, blockEvent.Timestamp)
	// if err == nil {
	// 	blockEvent.Timestamp = parsedTime.Format("January 2, 2006 15:04:05 MST")
	// }

	// // Update the currentEvent field in the Indexer
	// i.blockEvent = blockEvent

	// // Print the block event struct to the console
	// fmt.Printf("Received Event: %+v\n", blockEvent)

	// // Send the block event to the WebSocket clients
	// events <- blockEvent

	return nil
}

// Restart the Snek pipeline with the new node address
func (i *Indexer) Restart() {

	if i.isRunning {
		// Stop the current pipeline
		if err := i.pipeline.Stop(); err != nil {
			log.Fatalf("failed to stop pipeline: %s\n", err)
			log.Printf("failed to stop pipeline: %s\n", err)
			// Wait for a moment to ensure pipeline is fully stopped
			time.Sleep(time.Second)
		}
		i.isRunning = false
	}
	// Start a new pipeline with the updated node address
	if err := i.Start(); err != nil {
		log.Fatalf("failed to start pipeline: %s\n", err)
	}
}

// HTTP handler for updating the node address
func updateNodeAddressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newNodeAddress string
	if err := json.NewDecoder(r.Body).Decode(&newNodeAddress); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Check and update the node address
	if newNodeAddress == "" {
		newNodeAddress = "backbone.cardano-mainnet.iohk.io:3001" // Fallback to default
	}

	// Update the node address and restart the pipeline
	globalIndexer.nodeAddress = newNodeAddress
	globalIndexer.Restart()

	// Update the node address
	nodeMu.Lock()
	nodeMu.Unlock()

	// After updating the node address, send a message to all clients
	clientsMu.Lock()
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte("refresh"))
		if err != nil {
			log.Println(err)
			client.Close()
			delete(clients, client)
		}
	}
	clientsMu.Unlock()

	// Restart the Snek pipeline with the new node address
	globalIndexer.Restart()
	fmt.Printf("Updated node address to %s\n", globalIndexer.nodeAddress)
	// Refresh the web page
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func getNodeAddressHandler(w http.ResponseWriter, r *http.Request) {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	response := struct {
		NodeAddress string `json:"nodeAddress"`
	}{
		NodeAddress: globalIndexer.nodeAddress,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateEventTypeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newEventType string
	if err := json.NewDecoder(r.Body).Decode(&newEventType); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Update the event type and restart the pipeline
	globalIndexer.eventType = newEventType
	globalIndexer.Restart()

	fmt.Fprintf(w, "Event type updated to %s\n", newEventType)
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
	http.HandleFunc("/updateNodeAddress", updateNodeAddressHandler)
	http.HandleFunc("/getNodeAddress", getNodeAddressHandler)
	http.HandleFunc("/updateEventType", updateEventTypeHandler)

	// Start the HTTP server on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start HTTP server: %s", err)
	}
}
