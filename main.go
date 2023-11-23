package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/blinklabs-io/snek/event"
	filter_event "github.com/blinklabs-io/snek/filter/event"
	"github.com/blinklabs-io/snek/input/chainsync"
	output_embedded "github.com/blinklabs-io/snek/output/embedded"
	"github.com/blinklabs-io/snek/pipeline"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

}

// Ride the Snek

type Indexer struct {
	pipeline *pipeline.Pipeline
}

// Singleton indexer instance
var globalIndexer = &Indexer{}

// Define input options
var inputOpts = []chainsync.ChainSyncOptionFunc{
	chainsync.WithAddress("backbone.cardano-mainnet.iohk.io:3001"),
	chainsync.WithNetworkMagic(764824073),
	chainsync.WithIntersectTip(true),
}

func (i *Indexer) Start() error {
	// Create pipeline
	i.pipeline = pipeline.New()

	// Configure pipeline input
	input_chainsync := chainsync.New(
		inputOpts...,
	)

	i.pipeline.AddInput(input_chainsync)

	// Configure pipeline filters
	// We only care about block events
	filterEvent := filter_event.New(
		filter_event.WithTypes([]string{"chainsync.block"}),
	)
	i.pipeline.AddFilter(filterEvent)

	// Configure pipeline output
	output := output_embedded.New(
		output_embedded.WithCallbackFunc(i.handleEvent),
	)
	i.pipeline.AddOutput(output)

	// Start pipeline
	if err := i.pipeline.Start(); err != nil {
		log.Fatalf("failed to start pipeline: %s\n", err)
	}

	// Start error handler
	go func() {
		err, ok := <-i.pipeline.ErrorChan()
		if ok {
			log.Fatalf("pipeline failed: %s\n", err)
		}
	}()

	return nil
}

// Define handleEvent function
func (i *Indexer) handleEvent(event event.Event) error {

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	stringData := string(data)

	// Using the Struct we defined above to parse the event
	var blockEvent BlockEvent
	err = json.Unmarshal([]byte(stringData), &blockEvent)
	if err != nil {
		return err
	}

	// Print the block number
	fmt.Println(blockEvent.Context.BlockNumber)

	// Handle the event with the payload
	fmt.Println("Received event:", stringData)
	return nil
}

func main() {

	// Start snek
	if err := globalIndexer.Start(); err != nil {
		log.Fatalf("failed to start snek: %s", err)
	}

	// Rest of your code
	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)

	http.ListenAndServe(":8080", nil)
}
