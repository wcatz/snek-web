package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
)

type WebhookPayload struct {
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

var templates *template.Template
var mu sync.Mutex
var latestEventData WebhookPayload

func init() {
	templatesPath := filepath.Join(".", "templates", "*.html")
	templates = template.Must(template.ParseGlob(templatesPath))
}

func renderIndex(w http.ResponseWriter, templateName string, data interface{}) {
	err := templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func renderWebhookData(w http.ResponseWriter, templateName string, data interface{}) {
	err := templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Serve the index.html file with the latest event data
	renderIndex(w, "index.html", latestEventData)
}

func WebhookDataHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Serve the webhook_data.html file with the latest event data
	renderWebhookData(w, "webhook_data.html", latestEventData)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Unmarshal the JSON payload into a struct
	var payload WebhookPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "Error decoding JSON payload", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Update the latest event data
	latestEventData = payload
	renderIndex(w, "index.html", latestEventData)
}

func startSnek() {
	cmd := exec.Command("snek",
		"-input-chainsync-address", "m2:6002",
		"-filter-type", "chainsync.block",
		"-output", "webhook",
		"-output-webhook-url", "http://localhost:42069/webhook")

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error starting snek:", err)
		return
	}

	fmt.Println("Snek started successfully. PID:", cmd.Process.Pid)
}

func main() {
	// Start Snek asynchronously
	go startSnek()

	// Set up the Gorilla Mux router for the main server on port 8080
	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", HomeHandler).Methods("GET") // Handle root URL
	mainRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	mainRouter.HandleFunc("/webhook_data", WebhookDataHandler).Methods("GET") // New endpoint for webhook_data

	// Start the main HTTP server on port 8080
	mainPort := 8080
	mainAddr := fmt.Sprintf(":%v", mainPort)
	fmt.Printf("HTML server running on port %v...\n", mainPort)

	go func() {
		err := http.ListenAndServe(mainAddr, mainRouter)
		if err != nil {
			fmt.Println("Error starting main HTTP server:", err)
		}
	}()

	// Set up the Gorilla Mux router for the webhook server on port 42069
	webhookRouter := mux.NewRouter()
	webhookRouter.HandleFunc("/webhook", handleWebhook).Methods(http.MethodPost)

	// Start the webhook HTTP server on port 42069
	webhookPort := 42069
	webhookAddr := fmt.Sprintf(":%v", webhookPort)
	fmt.Printf("Webhook server running on port %v...\n", webhookPort)

	err := http.ListenAndServe(webhookAddr, webhookRouter)
	if err != nil {
		fmt.Println("Error starting webhook HTTP server:", err)
	}
}
