package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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

// Client represents a chat client
type Client struct {
	conn     *websocket.Conn
	nickname string
	room     string
	done     chan struct{}
	mu       sync.Mutex
}

// NewClient creates a new chat client
func NewClient() *Client {
	return &Client{
		done: make(chan struct{}),
	}
}

// Connect establishes a connection to the chat server
func (c *Client) Connect(url string) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	c.conn = conn
	return nil
}

// Login handles the client login process
func (c *Client) Login() error {
	var nickname string
	fmt.Print("Please enter your nickname: ")
	fmt.Scanln(&nickname)

	msg := Message{
		Author: nickname,
		Kind:   "init",
	}

	if err := c.sendMessage(msg); err != nil {
		return fmt.Errorf("failed to send login message: %v", err)
	}

	response, err := c.readMessage()
	if err != nil {
		return fmt.Errorf("failed to read login response: %v", err)
	}

	if !response.Mood {
		return fmt.Errorf("login failed: %s", response.Content)
	}

	c.nickname = nickname
	return nil
}

// JoinRoom handles the room joining process
func (c *Client) JoinRoom() error {
	fmt.Println("\nDo you want to:")
	fmt.Println("1. Create a new room")
	fmt.Println("2. Join an existing room")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		return c.createNewRoom()
	case "2":
		return c.joinExistingRoom()
	default:
		return fmt.Errorf("invalid choice")
	}
}

// createNewRoom handles creation of a new chat room
func (c *Client) createNewRoom() error {
	var roomName string
	fmt.Print("Enter room name: ")
	fmt.Scanln(&roomName)

	// First negotiate the room name
	msg := Message{
		Content: roomName,
		Kind:    "negotiate_name",
	}

	if err := c.sendMessage(msg); err != nil {
		return err
	}

	response, err := c.readMessage()
	if err != nil {
		return err
	}

	if !response.Mood {
		return fmt.Errorf("room name already exists")
	}

	// Create the room
	msg = Message{
		Author:  c.nickname,
		Content: roomName,
		Kind:    "new_room",
	}

	if err := c.sendMessage(msg); err != nil {
		return err
	}

	response, err = c.readMessage()
	if err != nil {
		return err
	}

	if !response.Mood {
		return fmt.Errorf("failed to create room")
	}

	c.room = roomName
	return nil
}

// joinExistingRoom handles joining an existing chat room
func (c *Client) joinExistingRoom() error {
	// For simplicity, just ask for room name directly
	var roomName string
	fmt.Print("Enter room name to join: ")
	fmt.Scanln(&roomName)

	msg := Message{
		Author:  c.nickname,
		Content: roomName,
		Kind:    "add_to_room",
	}

	if err := c.sendMessage(msg); err != nil {
		return err
	}

	response, err := c.readMessage()
	if err != nil {
		return err
	}

	if !response.Mood {
		return fmt.Errorf("failed to join room")
	}

	c.room = roomName
	return nil
}

// StartChat starts the chat session
func (c *Client) StartChat() error {
	// Start goroutine for receiving messages
	go c.receiveMessages()

	// Start goroutine for sending messages
	go c.sendMessages()

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	// Cleanup
	close(c.done)
	return c.conn.Close()
}

// receiveMessages handles incoming messages
func (c *Client) receiveMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			msg, err := c.readMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			if msg.Kind == "normal" {
				fmt.Printf("\nFrom %s in room %s: %s\n", msg.Author, msg.TargetedRoom, msg.Content)
			}
		}
	}
}

// sendMessages handles outgoing messages
func (c *Client) sendMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			var content string
			fmt.Print("\nEnter message (or 'quit' to exit): ")
			fmt.Scanln(&content)

			if content == "quit" {
				close(c.done)
				return
			}

			msg := Message{
				Author:       c.nickname,
				Content:      content,
				Timestamp:    time.Now().Format(time.RFC3339),
				Mood:         true,
				TargetedRoom: c.room,
				Kind:         "normal",
			}

			if err := c.sendMessage(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}

// helper functions
func (c *Client) sendMessage(msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(msg)
}

func (c *Client) readMessage() (Message, error) {
	var msg Message
	err := c.conn.ReadJSON(&msg)
	return msg, err
}

func main() {
	client := NewClient()

	// Connect to server
	if err := client.Connect("ws://localhost:8080/ws"); err != nil {
		log.Fatal(err)
	}

	// Login
	if err := client.Login(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully logged in as:", client.nickname)

	// Join room
	if err := client.JoinRoom(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully joined room:", client.room)

	// Start chat
	fmt.Println("Starting chat... (Press Ctrl+C to exit)")
	if err := client.StartChat(); err != nil {
		log.Fatal(err)
	}
}
