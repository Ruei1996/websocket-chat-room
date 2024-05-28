package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

// Define a WebSocket upgrader
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all connections
    },
}

var clients = make(map[*websocket.Conn]bool) // Store connected clients
var broadcast = make(chan Message) // Channel to queue messages to broadcast
var mutex = sync.Mutex{} // Mutex to protect access to shared resources

// Define the structure of a message
type Message struct {
    Username string `json:"username"` // User who sent the message
    Content  string `json:"content"`  // Content of the message
}

// Store chat history
var chatHistory []Message

// Handle new WebSocket connections
func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil) // Upgrade the HTTP connection to a WebSocket connection
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    mutex.Lock()
    clients[ws] = true // Add the new client
    mutex.Unlock()

    // Listen for new messages from this client
    for {
        var msg Message
        err := ws.ReadJSON(&msg) // Read a new message as JSON
        if err != nil {
            log.Printf("error: %v", err)
            mutex.Lock()
            delete(clients, ws) // Remove the client
            mutex.Unlock()
            break
        }
        broadcast <- msg // Send the new message to the broadcast channel
        chatHistory = append(chatHistory, msg) // Store the message in the chat history
    }
}

// Handle messages to broadcast to all clients
func handleMessages() {
    for {
        msg := <-broadcast // Wait for a new message to broadcast
        mutex.Lock()
        for client := range clients {
            err := client.WriteJSON(msg) // Send the message to the client
            if err != nil {
                log.Printf("error: %v", err)
                client.Close()
                delete(clients, client) // Remove the client
            }
        }
        mutex.Unlock()
    }
}

// Handle requests to dump the chat history
func handleDump(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()

    response := map[string]interface{}{
        "chatHistory": chatHistory, // Include the chat history in the response
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response) // Send the response as JSON
}

func main() {
    fs := http.FileServer(http.Dir("./public")) // Serve static files from the "public" directory
    http.Handle("/", fs) // Handle requests to the root path ("/") with the file server
    http.HandleFunc("/ws", handleConnections) // Handle WebSocket connections at the "/ws" path
    http.HandleFunc("/dump", handleDump) // Handle requests to dump the chat history at the "/dump" path

    go handleMessages() // Start handling messages

    log.Println("http server started on :8080")
    err := http.ListenAndServe(":8080", nil) // Start the HTTP server on port 8080
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}