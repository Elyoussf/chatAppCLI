package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"websocket/ui"

	"github.com/gorilla/websocket"
)

type ChatClient struct {
	messages []ui.Message
	lock     sync.Mutex
	id       int
}

func NewChatClient() *ChatClient {
	return &ChatClient{
		messages: make([]ui.Message, 0),
	}
}

func (cc *ChatClient) AddMessage(content, author string) {
	cc.lock.Lock()
	defer cc.lock.Unlock()
	cc.id++
	cc.messages = append(cc.messages, ui.Message{
		Content: content,
		Author:  author,
		ID:      cc.id,
	})
}

func (cc *ChatClient) GetMessages() []ui.Message {
	cc.lock.Lock()
	defer cc.lock.Unlock()
	return append([]ui.Message(nil), cc.messages...)
}

func HandleClientConnection(serverURL string, client *ChatClient) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine for sending messages

	go func() {
		defer wg.Done()
		for {
			input := ui.GetUserInput()
			client.AddMessage(input, "Client")
			if err := conn.WriteMessage(websocket.TextMessage, []byte(input)); err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
			ui.DrawTerminal(client.GetMessages())
		}
	}()

	// Goroutine for receiving messages
	go func() {
		defer wg.Done()
		for {
			ui.DrawTerminal(client.GetMessages())
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				return
			}
			client.AddMessage(string(msg), "Server")

		}
	}()
	fmt.Println("If you wanna initiate the communication start typing it and hit Enter key")

	wg.Wait()
}

func main() {

	client := NewChatClient()
	serverURL := "ws://localhost:8080/ws"

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		HandleClientConnection(serverURL, client)
	}()

	<-interrupt
	log.Println("Client shutting down...")
}
