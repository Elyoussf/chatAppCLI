package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"websocket/ui"

	"github.com/gorilla/websocket"
)

type ChatServer struct {
	messages []ui.Message
	lock     sync.Mutex
	id       int
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		messages: make([]ui.Message, 0),
	}
}

func (cs *ChatServer) AddMessage(content, author string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	cs.id++
	cs.messages = append(cs.messages, ui.Message{
		Content: content,
		Author:  author,
		ID:      cs.id,
	})
}

func (cs *ChatServer) GetMessages() []ui.Message {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	return append([]ui.Message(nil), cs.messages...)
}

func HandleServerConnection(conn *websocket.Conn, server *ChatServer) {
	defer conn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine for sending messages
	go func() {
		defer wg.Done()
		for {
			ui.DrawTerminal(server.GetMessages())
			input := ui.GetUserInput()
			server.AddMessage(input, "Server")
			if err := conn.WriteMessage(websocket.TextMessage, []byte(input)); err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
			ui.DrawTerminal(server.GetMessages())
		}
	}()

	// Goroutine for receiving messages
	go func() {
		defer wg.Done()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				return
			}
			server.AddMessage(string(msg), "Client")
			ui.DrawTerminal(server.GetMessages())
		}
	}()

	wg.Wait()
}

func main() {
	server := NewChatServer()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r, http.Header{}, 0, 0)
		if err != nil {
			log.Printf("Failed to upgrade: %v", err)
			return
		}
		HandleServerConnection(conn, server)
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	log.Println("Server is running on ws://localhost:8080/ws")
	<-interrupt
	log.Println("Server shutting down...")
}
