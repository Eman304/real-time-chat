package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"time"
)

// Message structure matching server
type Message struct {
	ClientID string
	Content  string
}

// JoinRequest structure
type JoinRequest struct {
	ClientID string
}

// JoinResponse structure
type JoinResponse struct {
	ClientID string
}

// Client ID generator
var (
	clientID string
	mu       sync.Mutex
	quit     chan bool
)

// listenForMessages continuously polls server for new messages
func listenForMessages(client *rpc.Client, clientID string) {
	for {
		select {
		case <-quit:
			return
		default:
		}

		var messages []string
		err := client.Call("ChatServer.GetMessages", clientID, &messages)
		if err != nil {
			log.Printf("❌ Error receiving message: %v", err)
			continue
		}

		// Display received message
		if len(messages) > 0 {
			fmt.Printf("\n%s\n", messages[0])
			fmt.Print("You: ")
		}
	}
}

// handleUserInput reads user input and sends messages
func handleUserInput(client *rpc.Client, clientID string, reader *bufio.Reader) {
	for {
		fmt.Print("You: ")
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)

		if msg == "" {
			continue
		}

		if msg == "exit" {
			fmt.Println("👋 Goodbye!")
			// Notify server of departure
			var res JoinResponse
			client.Call("ChatServer.Leave", &JoinRequest{ClientID: clientID}, &res)
			quit <- true
			break
		}

		// Send message to server
		message := &Message{
			ClientID: clientID,
			Content:  msg,
		}
		var response Message
		err := client.Call("ChatServer.SendMessage", message, &response)
		if err != nil {
			log.Printf("⚠️ Error sending message: %v", err)
		}
	}
}

func main() {
	// Generate unique client ID
	clientID = fmt.Sprintf("User_%d", time.Now().UnixNano()%100000)
	quit = make(chan bool)

	// Connect to server
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("❌ Error connecting to server:", err)
	}
	defer client.Close()

	// Join the chatroom
	var res JoinResponse
	err = client.Call("ChatServer.Join", &JoinRequest{ClientID: clientID}, &res)
	if err != nil {
		log.Fatal("❌ Error joining chatroom:", err)
	}

	fmt.Println("✅ Connected to chatroom!")
	fmt.Printf("📝 Your ID: %s\n", clientID)
	fmt.Println("Type your message below (type 'exit' to quit):\n")

	// Start goroutine to listen for incoming messages
	go listenForMessages(client, clientID)

	// Handle user input in main goroutine
	reader := bufio.NewReader(os.Stdin)
	handleUserInput(client, clientID, reader)
}
