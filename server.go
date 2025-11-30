package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

// Client represents a connected client
type Client struct {
	id       string
	conn     net.Conn
	sendChan chan string
}

// ChatServer manages all connected clients and broadcasts
type ChatServer struct {
	mu      sync.Mutex      // Protects clients map
	clients map[string]*Client
	nextID  int
}

// Message structure for RPC communication
type Message struct {
	ClientID string
	Content  string
}

// JoinRequest represents a client joining the chat
type JoinRequest struct {
	ClientID string
}

// JoinResponse confirms the join
type JoinResponse struct {
	ClientID string
}

// Join RPC method - called when client connects
func (c *ChatServer) Join(req *JoinRequest, res *JoinResponse) error {
	c.mu.Lock()
	client := &Client{
		id:       req.ClientID,
		sendChan: make(chan string, 10), // buffered channel
	}
	c.clients[req.ClientID] = client
	numClients := len(c.clients)
	c.mu.Unlock()

	res.ClientID = req.ClientID
	log.Printf("✅ User %s joined! Total clients: %d", req.ClientID, numClients)

	// Broadcast join notification to all other clients
	joinMsg := fmt.Sprintf("📢 User %s joined", req.ClientID)
	c.broadcastToOthers(req.ClientID, joinMsg)

	return nil
}

// SendMessage RPC method - called when client sends a message
func (c *ChatServer) SendMessage(msg *Message, res *Message) error {
	log.Printf("📨 Message from %s: %s", msg.ClientID, msg.Content)

	// Broadcast to all other clients
	broadcastMsg := fmt.Sprintf("[%s]: %s", msg.ClientID, msg.Content)
	c.broadcastToOthers(msg.ClientID, broadcastMsg)

	res.ClientID = msg.ClientID
	res.Content = "message_received"
	return nil
}

// Leave RPC method - called when client disconnects
func (c *ChatServer) Leave(req *JoinRequest, res *JoinResponse) error {
	c.mu.Lock()
	if client, exists := c.clients[req.ClientID]; exists {
		delete(c.clients, req.ClientID)
		close(client.sendChan)
	}
	numClients := len(c.clients)
	c.mu.Unlock()

	res.ClientID = req.ClientID
	log.Printf("👋 User %s left! Total clients: %d", req.ClientID, numClients)

	// Broadcast leave notification to all other clients
	leaveMsg := fmt.Sprintf("📢 User %s left", req.ClientID)
	c.broadcastToOthers(req.ClientID, leaveMsg)

	return nil
}

// GetMessages RPC method - for client to receive broadcast messages
func (c *ChatServer) GetMessages(clientID *string, messages *[]string) error {
	c.mu.Lock()
	client, exists := c.clients[*clientID]
	c.mu.Unlock()

	if !exists {
		return fmt.Errorf("client %s not found", *clientID)
	}

	// Wait for messages on the channel and collect them
	// This will block until there's a message to send
	msg := <-client.sendChan
	*messages = []string{msg}

	return nil
}

// broadcastToOthers sends a message to all clients except the sender
func (c *ChatServer) broadcastToOthers(senderID string, msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, client := range c.clients {
		if id != senderID { // Don't send to self
			select {
			case client.sendChan <- msg:
				// Message sent
			default:
				log.Printf("⚠️ Warning: channel full for client %s", id)
			}
		}
	}
}

func main() {
	// Create chat server
	chatServer := &ChatServer{
		clients: make(map[string]*Client),
		nextID:  1,
	}

	// Register RPC service
	rpc.Register(chatServer)

	// Start listening for TCP connections
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	log.Println("🚀 Real-time Chat Server is running on port 1234...")
	log.Println("Waiting for clients to connect...")

	// Accept client connections continuously
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		// Handle each client connection in a separate goroutine
		go rpc.ServeConn(conn)
	}
}
