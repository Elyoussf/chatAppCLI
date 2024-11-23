package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a chat message
type Message struct {
	Author       string `json:"author"`
	Content      string `json:"content"`
	Timestamp    string `json:"timestamp"`
	Mood         bool   `json:"mood"`
	TargetedRoom string `json:"targetedRoom"`
	Kind         string `json:"kind"`
}

// Client represents a connected client
type Client struct {
	conn     *websocket.Conn
	nickname string
	rooms    map[string]bool
	mu       sync.Mutex
}

// Room represents a chat room
type Room struct {
	name    string
	clients map[string]*Client
	mu      sync.RWMutex
}

// Server represents the chat server
type Server struct {
	clients    map[string]*Client
	rooms      map[string]*Room
	upgrader   websocket.Upgrader
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

// NewServer creates a new chat server instance
func NewServer() *Server {
	return &Server{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for testing
			},
		},
	}
}

// Start starts the chat server
func (s *Server) Start() {
	// Handle WebSocket connections
	http.HandleFunc("/ws", s.handleConnection)

	// Start the main server loop
	go s.run()

	// Start HTTP server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// run processes client registration and unregistration
func (s *Server) run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.nickname] = client
			s.mu.Unlock()
			log.Printf("Client registered: %s", client.nickname)

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client.nickname]; ok {
				delete(s.clients, client.nickname)
				client.conn.Close()
			}
			s.mu.Unlock()
			log.Printf("Client unregistered: %s", client.nickname)

			// Remove client from all rooms
			s.mu.RLock()
			for _, room := range s.rooms {
				room.mu.Lock()
				delete(room.clients, client.nickname)
				room.mu.Unlock()
			}
			s.mu.RUnlock()
		}
	}
}

// handleConnection handles incoming WebSocket connections
func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}

	client := &Client{
		conn:  conn,
		rooms: make(map[string]bool),
	}

	// Handle client messages
	go s.handleClient(client)
}

// handleClient processes messages from a connected client
func (s *Server) handleClient(client *Client) {
	defer func() {
		s.unregister <- client
	}()

	for {
		var msg Message
		err := client.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		switch msg.Kind {
		case "init":
			s.handleInitMessage(client, msg)
		case "negotiate_name":
			s.handleNegotiateRoom(client, msg)
		case "new_room":
			s.handleNewRoom(client, msg)
		case "add_to_room":
			s.handleJoinRoom(client, msg)
		case "normal":
			s.handleChatMessage(client, msg)
		}
	}
}

// handleInitMessage handles client initialization
func (s *Server) handleInitMessage(client *Client, msg Message) {
	s.mu.RLock()
	_, exists := s.clients[msg.Author]
	s.mu.RUnlock()

	response := Message{
		Kind: "init",
		Mood: !exists,
	}

	if exists {
		response.Content = "Nickname already taken"
		client.conn.WriteJSON(response)
		return
	}

	client.nickname = msg.Author
	s.register <- client
	client.conn.WriteJSON(response)
}

// handleNegotiateRoom handles room name negotiation
func (s *Server) handleNegotiateRoom(client *Client, msg Message) {
	s.mu.RLock()
	_, exists := s.rooms[msg.Content]
	s.mu.RUnlock()

	response := Message{
		Kind:    "negotiate_name",
		Mood:    !exists,
		Content: msg.Content,
	}

	client.conn.WriteJSON(response)
}

// handleNewRoom handles creation of new rooms
func (s *Server) handleNewRoom(client *Client, msg Message) {
	s.mu.Lock()
	room := &Room{
		name:    msg.Content,
		clients: make(map[string]*Client),
	}
	s.rooms[msg.Content] = room
	s.mu.Unlock()

	room.mu.Lock()
	room.clients[client.nickname] = client
	room.mu.Unlock()

	client.mu.Lock()
	client.rooms[msg.Content] = true
	client.mu.Unlock()

	response := Message{
		Kind:    "new_room",
		Content: msg.Content,
		Mood:    true,
	}

	client.conn.WriteJSON(response)
}

// handleJoinRoom handles clients joining existing rooms
func (s *Server) handleJoinRoom(client *Client, msg Message) {
	s.mu.RLock()
	room, exists := s.rooms[msg.Content]
	s.mu.RUnlock()

	response := Message{
		Kind:    "add_to_room",
		Content: msg.Content,
		Mood:    exists,
	}

	if !exists {
		client.conn.WriteJSON(response)
		return
	}

	room.mu.Lock()
	room.clients[client.nickname] = client
	room.mu.Unlock()

	client.mu.Lock()
	client.rooms[msg.Content] = true
	client.mu.Unlock()

	client.conn.WriteJSON(response)
}

// handleChatMessage handles normal chat messages
func (s *Server) handleChatMessage(client *Client, msg Message) {
	msg.Timestamp = time.Now().Format(time.RFC3339)

	s.mu.RLock()
	room, exists := s.rooms[msg.TargetedRoom]
	s.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	for _, targetClient := range room.clients {
		targetClient.mu.Lock()
		targetClient.conn.WriteJSON(msg)
		targetClient.mu.Unlock()
	}
	room.mu.RUnlock()
}

func main() {
	server := NewServer()
	server.Start()
}
