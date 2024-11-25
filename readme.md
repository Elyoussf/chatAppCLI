# Real-Time Chat Application with WebSocket using Go 🚀

A real-time chat application implemented in Go using WebSocket protocol for seamless communication between multiple clients and a central server.

## Features ✨

- Real-time messaging between multiple clients
- Simple and interactive command-line interface
- Built with native Go WebSocket implementation
- Lightweight and efficient communication

## Prerequisites 📋

Before running the application, ensure you have:
- Go installed on your system (version 1.16 or later recommended)
- Basic understanding of terminal/command-line operations

## Getting Started 🌟

### Starting the Server

First, start the server by running:

```bash
go run Server/exec/main.go
```

The server will start and listen for incoming WebSocket connections.

### Creating Chat Clients

You can create as many chat clients as you want by running:

```bash
go run Client/exec/main.go
```

Each client instance will provide an interactive command-line interface for chatting.

## How It Works 🔄

1. The server starts and waits for client connections
2. Each client connects to the server via WebSocket
3. Clients can send messages that will be broadcasted to all connected clients
4. The communication happens in real-time with minimal latency

## Project Structure 📁

```
.
├── Server/
│   └── exec/
│       └── main.go    # Server implementation
└── Client/
    └── exec/
        └── main.go    # Client implementation
```

## Contributing 🤝

Feel free to fork this repository and submit pull requests. For major changes, please open an issue first to discuss what you would like to change.

## License 📄

Ait Elghazi license
