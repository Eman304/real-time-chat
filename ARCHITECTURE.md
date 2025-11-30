# Architecture & Design Documentation

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     REAL-TIME CHATROOM SYSTEM                   │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                          SERVER (port 1234)                      │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ ChatServer (Protected by Mutex)                            │ │
│  │                                                            │ │
│  │  clients map[string]*Client                               │ │
│  │  ├─ "User_123" → *Client                                  │ │
│  │  ├─ "User_456" → *Client                                  │ │
│  │  └─ "User_789" → *Client                                  │ │
│  │                                                            │ │
│  │  RPC Methods:                                             │ │
│  │  • Join(clientID) → broadcast join notification           │ │
│  │  • SendMessage(msg) → broadcast to others                 │ │
│  │  • Leave(clientID) → broadcast leave notification         │ │
│  │  • GetMessages() → blocking call to receive from channel  │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Client 1 Structure                                       │   │
│  │ ├─ id: "User_123"                                        │   │
│  │ ├─ conn: net.Conn                                        │   │
│  │ └─ sendChan: chan string (buffered, size 10)             │   │
│  │    └─ Receives: "[User_456]: Hello"                      │   │
│  │    └─ Receives: "📢 User_456 joined"                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Client 2 Structure                                       │   │
│  │ ├─ id: "User_456"                                        │   │
│  │ ├─ conn: net.Conn                                        │   │
│  │ └─ sendChan: chan string (buffered, size 10)             │   │
│  │    └─ Receives: "[User_123]: How are you?"               │   │
│  │    └─ Receives: "📢 User_789 joined"                     │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                      CLIENT 1 (Go Goroutines)                    │
│                                                                  │
│  ┌──────────────────────────┐      ┌──────────────────────────┐ │
│  │ Goroutine 1:             │      │ Main Thread:             │ │
│  │ listenForMessages()      │      │ handleUserInput()        │ │
│  │                          │      │                          │ │
│  │ for {                    │      │ for {                    │ │
│  │   msg := <-client.       │      │   input = readInput()    │ │
│  │           GetMessages()  │      │   if input == "exit" {   │ │
│  │   fmt.Print(msg)         │      │     Leave()              │ │
│  │ }                        │      │     break                │ │
│  │                          │      │   }                      │ │
│  │ Blocks until message     │      │   SendMessage(input)     │ │
│  │ arrives from server      │      │ }                        │ │
│  └──────────────────────────┘      └──────────────────────────┘ │
│         ↓ (receives from channel)        ↓ (sends via RPC)      │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                      CLIENT 2 (Go Goroutines)                    │
│                                                                  │
│  ┌──────────────────────────┐      ┌──────────────────────────┐ │
│  │ Goroutine 1:             │      │ Main Thread:             │ │
│  │ listenForMessages()      │      │ handleUserInput()        │ │
│  │ (Same structure)         │      │ (Same structure)         │ │
│  └──────────────────────────┘      └──────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

## Message Flow Sequence

### Scenario: Client1 sends "Hello" and Client2 receives it

```
Timeline:
─────────────────────────────────────────────────────────────────────

T1: CLIENT 1 (main thread)
    "You: Hello"
    → Calls SendMessage(Message{ClientID: "User_123", Content: "Hello"})

T2: SERVER (RPC Goroutine)
    ← Receives SendMessage RPC call
    → Acquires Mutex lock
    → Gets Client map
    → Broadcasts to all except "User_123"
    → Sends "[User_123]: Hello" to Client2's sendChan
    → Releases Mutex lock

T3: CLIENT 2 (Message Listener Goroutine)
    ← Receives "[User_123]: Hello" from channel
    → Displays: "[User_123]: Hello"
    → Calls GetMessages() again to wait for next message

T4: CLIENT 1 (Message Listener Goroutine)
    ← Still waiting on GetMessages()
    ← Does NOT receive own message (broadcast filters sender)
```

## Data Structure and Synchronization

### Mutex Protected Section

```go
// CRITICAL SECTION (protected by Mutex)
c.mu.Lock()
{
    for id, client := range c.clients {
        if id != senderID {
            select {
            case client.sendChan <- msg:
                // Non-blocking send
            default:
                log.Printf("Channel full")
            }
        }
    }
}
c.mu.Unlock()
```

**Why Mutex?**
- Prevents concurrent map iteration errors
- Prevents client removal during iteration
- Ensures consistent snapshot of clients

**Lock Duration?**
- Minimal: only while iterating clients
- Released before blocking on channel sends
- Released before waiting for RPC responses

### Channel Operation Diagram

```
Client 1 sendChan (buffer size: 10)
┌─────────────────────────┐
│ Message Queue           │
├─────────────────────────┤
│ [1] "[User_2]: Hi"      │ ← Most recent
│ [2] "📢 User_3 joined"  │
│ [3] (empty)             │
│ ...                     │
└─────────────────────────┘
     ↑
     │ Server sends here (select/default)
     │
     └─ Client reader waits here
        (blocks until message arrives)
```

## Concurrency Flow Chart

```
┌─────────────────────────────────────────────────────────────┐
│ Client joins:                                               │
│                                                             │
│ 1. Client calls Join RPC                                    │
│ 2. Server acquires Mutex                                    │
│ 3. Server adds client to map                                │
│ 4. Server releases Mutex                                    │
│ 5. Server broadcasts "User_X joined" to all others          │
│ 6. All other clients receive via channel                    │
│ 7. All other clients display notification                   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Client sends message:                                       │
│                                                             │
│ 1. Client calls SendMessage RPC                             │
│ 2. Server receives in goroutine                             │
│ 3. Server acquires Mutex                                    │
│ 4. Server iterates client map                               │
│ 5. For each other client:                                   │
│    5a. Non-blocking send to channel (select/default)        │
│    5b. If channel full, log warning                         │
│ 6. Server releases Mutex                                    │
│ 7. All other clients' listeners unblock                     │
│ 8. Messages displayed on all clients                        │
│ 9. Listeners call GetMessages() again                       │
│ 10. Listeners block until next message                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Client disconnects:                                         │
│                                                             │
│ 1. Client calls Leave RPC or closes connection              │
│ 2. Server receives Leave RPC in goroutine                   │
│ 3. Server acquires Mutex                                    │
│ 4. Server removes client from map                           │
│ 5. Server closes client's channel                           │
│ 6. Server releases Mutex                                    │
│ 7. Server broadcasts "User_X left" to remaining clients     │
│ 8. Remaining clients receive notification via channel       │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Why Buffered Channels (Size 10)?
- Prevents sender blocking if receiver temporarily slow
- Allows server to send multiple messages without waiting
- Balances memory usage vs. responsiveness
- If buffer full: warning logged, next message might be lost

### 2. Why Mutex Instead of RWMutex?
- Write operations (add/remove client) are frequent
- Lock contention is minimal (quick operations)
- Simpler implementation
- RWMutex adds complexity without benefit here

### 3. Why Two Separate RPC Calls?
- `SendMessage()`: For sending (non-blocking)
- `GetMessages()`: For receiving (blocking)
- Allows client to wait without polling
- Cleaner API separation

### 4. Why Non-Blocking Send with select/default?
- Server doesn't block on slow receivers
- One slow client doesn't affect others
- Prevents deadlocks with full channels
- Trade-off: might lose messages if buffer full

### 5. Why No Message History?
- Real-time broadcast philosophy
- Reduces memory usage
- Simpler implementation
- Late joiners see future messages, not past

## Thread Safety Analysis

### Race Conditions Prevented

1. **Concurrent Map Access**
   - Mutex protects `clients` map
   - No concurrent reads/writes to map

2. **Client Addition/Removal During Broadcast**
   - Mutex held during iteration
   - No map modification during broadcast

3. **Channel Send/Receive**
   - Channels are thread-safe primitives
   - No explicit synchronization needed

4. **Goroutine Leaks**
   - Each client goroutine has clean exit path
   - Message listener exits on `quit` signal
   - Server goroutines exit when client disconnects

### Memory Safety

- No shared memory without synchronization
- All channel accesses are synchronized
- Deferred mutex unlock prevents deadlocks
- Go runtime manages goroutine cleanup
