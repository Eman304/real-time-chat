# 💬 Real-Time Chatroom using Go RPC with Concurrency

## 📽️ Demo
🎥 [Recording Link](https://screenapp.io/app/v/ml6oCrEZrJ) #Archived demo recording — removed as outdated#

## 📘 Description
A real-time chatroom built in Go using RPC with concurrent message broadcasting.
- **Real-time Broadcasting**: Messages are instantly broadcast to all connected clients
- **Goroutines & Channels**: Uses Go's concurrency primitives for non-blocking I/O
- **Mutex Protection**: Shared client list is synchronized with `sync.Mutex`
- **Join/Leave Notifications**: All clients are notified when someone joins or leaves
- **No Self-Echo**: Clients don't receive their own messages

## 🏗️ Architecture

### Server (`server.go`)
- Manages a map of connected clients (protected by Mutex)
- RPC methods:
  - `Join()`: Client joins the chatroom, triggers join notification broadcast
  - `SendMessage()`: Client sends message, broadcasts to all others
  - `Leave()`: Client leaves, triggers leave notification broadcast
  - `GetMessages()`: Blocking RPC call to receive messages from channel

### Client (`client.go`)
- Two concurrent goroutines:
  1. **Message Listener**: Goroutine that blocks on `GetMessages()` RPC calls to receive broadcasts
  2. **Input Handler**: Main thread handles user input and sends messages
- Automatically generates unique client ID

## 🚀 How to Run

### 1️⃣ Run the server:
```bash
go run server.go
```

### 2️⃣ Run the client (in separate terminals):
```bash
go run client.go
```

Repeat step 2 for each client you want to connect.

### 3️⃣ Type messages and see real-time broadcasting:
- Messages appear on all clients instantly
- Join/leave notifications are broadcast to all
- Type `exit` to quit

## 🔧 Key Implementation Details

### Concurrency Model
- **Goroutines**: Each client connection is handled in a separate goroutine
- **Channels**: Buffered channels (size 10) pass messages between goroutines
- **Mutex**: Protects the shared client map from race conditions
- **Select Statement**: Used in client to handle graceful shutdown

### Message Flow
1. Client sends message via `SendMessage()` RPC call
2. Server receives RPC call in a goroutine
3. Server locks client map, iterates through all clients
4. For each other client, server sends message to their channel (non-blocking)
5. Client's listener goroutine receives from channel and displays message

### Synchronization
```
Client Map (Protected by Mutex)
│
├─ User_123 → Channel with pending messages
├─ User_456 → Channel with pending messages
└─ User_789 → Channel with pending messages
```

## 📊 Performance Characteristics
- **Scalability**: Can handle multiple concurrent clients
- **Responsiveness**: Messages delivered in real-time via channels
- **Safety**: No race conditions due to proper mutex usage
- **Blocking Design**: `GetMessages()` blocks until a message is available

## 🎯 Improvements Over Original RPC System
- ✅ Real-time message delivery (not pull-based history)
- ✅ Concurrent client handling with goroutines
- ✅ No self-echo (messages don't appear twice)
- ✅ Join/leave notifications for awareness
- ✅ Proper synchronization with mutex
- ✅ Buffered channels for efficient message passing
