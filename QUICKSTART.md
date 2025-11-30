# 🚀 Quick Start Guide

## 30-Second Setup

### 1. Start Server (Terminal 1)
```powershell
cd "c:\Users\EMAM ABD EL MONSEF\Desktop\Eman Monsef\4th year\distrbuted systems\simple-chatroom"
go run server.go
```

**Expected output:**
```
🚀 Real-time Chat Server is running on port 1234...
Waiting for clients to connect...
```

### 2. Start Client (Terminal 2)
```powershell
cd "c:\Users\EMAM ABD EL MONSEF\Desktop\Eman Monsef\4th year\distrbuted systems\simple-chatroom"
go run client.go
```

**Expected output:**
```
✅ Connected to chatroom!
📝 Your ID: User_XXXXX
Type your message below (type 'exit' to quit):

You:
```

### 3. Start More Clients (Terminal 3+)
```powershell
go run client.go
```

### 4. Test It!

**In Terminal 2 (first client):**
```
You: Hello everyone!
```

**Expected in Terminal 3+ (other clients):**
```
[User_12345]: Hello everyone!
```

---

## What You're Seeing

| Event | Display | Goroutines Running |
|-------|---------|-------------------|
| Client joins | "📢 User_X joined" on others | 2 (input + listener) |
| Client sends message | Message appears on all others | 2 per client |
| Client types "exit" | "📢 User_X left" on others | 0 (graceful shutdown) |

---

## Key Concurrency Points

1. **Server**: Main goroutine accepts connections, each client gets its own goroutine
2. **Client**: Two goroutines run concurrently:
   - Input handler (main thread)
   - Message listener (waits for broadcasts)

3. **Synchronization**:
   - Mutex protects client list
   - Channels pass messages

---

## Stopping Everything

- **Client**: Type `exit` and press Enter
- **Server**: Press `Ctrl+C` in Terminal 1

---

## Next: Push to GitHub

Follow `GITHUB_SETUP.md` to create a new repository and push this code.

---

## For More Details

- 📖 **Architecture**: See `ARCHITECTURE.md`
- 🧪 **Testing**: See `TESTING_GUIDE.md`
- 📝 **Full Docs**: See `README.md`
- 🚀 **GitHub**: See `GITHUB_SETUP.md`
- ✅ **Summary**: See `COMPLETION_SUMMARY.md`
