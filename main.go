package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/blinklabs-io/snek/event"
	filter_event "github.com/blinklabs-io/snek/filter/event"
	"github.com/blinklabs-io/snek/input/chainsync"
	output_embedded "github.com/blinklabs-io/snek/output/embedded"
	"github.com/blinklabs-io/snek/pipeline"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	events    = make(chan BlockEvent)
	templates *template.Template
)

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

type TemplateData struct {
	BlockEvent BlockEvent
}

func init() {
	templatesPath := filepath.Join(".", "templates", "*.html")
	templates = template.Must(template.ParseGlob(templatesPath))
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a sample BlockEvent (replace this with your actual data)
	blockEvent := BlockEvent{
		Type: "sample",
		// ... other fields ...
	}

	data := TemplateData{
		BlockEvent: blockEvent,
	}

	// Pass the TemplateData to the template
	tmpl.Execute(w, data)
}

type Indexer struct {
	pipeline   *pipeline.Pipeline
	blockEvent BlockEvent
}

var globalIndexer = &Indexer{}

var inputOpts = []chainsync.ChainSyncOptionFunc{
	chainsync.WithAddress("backbone.cardano-mainnet.iohk.io:3001"),
	chainsync.WithNetworkMagic(764824073),
	chainsync.WithIntersectTip(true),
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
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

func (i *Indexer) Start() error {
	i.pipeline = pipeline.New()

	input_chainsync := chainsync.New(
		inputOpts...,
	)

	i.pipeline.AddInput(input_chainsync)

	filterEvent := filter_event.New(
		filter_event.WithTypes([]string{"chainsync.block"}),
	)
	i.pipeline.AddFilter(filterEvent)

	output := output_embedded.New(
		output_embedded.WithCallbackFunc(i.handleEvent),
	)
	i.pipeline.AddOutput(output)

	if err := i.pipeline.Start(); err != nil {
		log.Fatalf("failed to start pipeline: %s\n", err)
	}

	go func() {
		err, ok := <-i.pipeline.ErrorChan()
		if ok {
			log.Fatalf("pipeline failed: %s\n", err)
		}
	}()

	return nil
}

func (i *Indexer) handleEvent(event event.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	stringData := string(data)

	var blockEvent BlockEvent
	err = json.Unmarshal([]byte(stringData), &blockEvent)
	if err != nil {
		return err
	}

	i.blockEvent = blockEvent

	fmt.Println(blockEvent.Context.BlockNumber)

	// Send the block event to the WebSocket clients
	events <- blockEvent

	return nil
}

func main() {
	if err := globalIndexer.Start(); err != nil {
		log.Fatalf("failed to start snek: %s", err)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)

	http.ListenAndServe(":8080", nil)
}
