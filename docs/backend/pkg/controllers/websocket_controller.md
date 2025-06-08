# WebSocket Controller Documentation

> This document describes the WebSocket controller that enables real-time bidirectional communication between clients and the NIVAI API using a hub-based publish/subscribe pattern.

## Architecture

```mermaid
classDiagram
    class Hub {
        -Map~Client,bool~ clients
        -Channel register
        -Channel unregister
        -Channel broadcast
        -Mutex mu
        +NewHub() Hub
        +Run()
    }

    class Client {
        -WebSocket conn
        -Channel send
        -Hub hub
        +readPump()
        +writePump()
    }

    class WebSocketHandler {
        +Handle(w, r)
        -upgrader WebSocketUpgrader
    }

    Hub "1" --> "*" Client : manages
    Client --> Hub : sends messages to
    WebSocketHandler ..> Client : creates
    WebSocketHandler ..> Hub : uses
```

## Message Flow

```mermaid
sequenceDiagram
    participant C1 as Client 1
    participant WS as WebSocket Handler
    participant H as Hub
    participant C2 as Client 2

    C1->>WS: HTTP Connection
    WS->>WS: Upgrade to WebSocket
    WS->>H: Register Client 1

    activate C1
    note over C1,H: Start Client Pumps

    C2->>WS: HTTP Connection
    WS->>WS: Upgrade to WebSocket
    WS->>H: Register Client 2

    C1->>H: Send Message
    H->>C1: Broadcast Message
    H->>C2: Broadcast Message

    C1-->>H: Disconnect
    H-->>H: Unregister Client 1
    deactivate C1
```

## Components

### Hub

Central message broker that:

```go
type Hub struct {
    clients    map[*Client]bool   // Active connections
    register   chan *Client       // Registration requests
    unregister chan *Client      // Unregistration requests
    broadcast  chan []byte        // Broadcast messages
    mu         sync.Mutex        // Thread safety
}
```

### Client

Represents a WebSocket connection:

```go
type Client struct {
    conn *websocket.Conn     // WebSocket connection
    send chan []byte         // Message buffer
    hub  *Hub               // Hub reference
}
```

## Connection Lifecycle

### 1. Connection Establishment

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler
    participant Hub as Hub

    C->>H: HTTP Request
    H->>H: Upgrade Connection
    H->>Hub: Register Client
    activate C
    par Read Pump
        C->>C: Start readPump()
    and Write Pump
        C->>C: Start writePump()
    end
```

### 2. Message Broadcasting

```mermaid
sequenceDiagram
    participant S as Sender
    participant H as Hub
    participant R1 as Receiver 1
    participant R2 as Receiver 2

    S->>H: Send Message
    par Broadcast
        H->>R1: Forward Message
    and
        H->>R2: Forward Message
    end
```

## Configuration

### WebSocket Settings

```go
upgrader := websocket.Upgrader{
    ReadBufferSize: 1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true  // Configure for production
    },
}
```

## Performance Features

### 1. Message Buffering

- 256 message buffer per client
- Non-blocking broadcast operations
- Automatic client cleanup on buffer overflow

### 2. Concurrency Management

- Thread-safe client management
- Goroutine-based message pumps
- Mutex-protected shared resources

## Error Handling

1. **Connection Errors**

   - Unexpected closure detection
   - Graceful connection termination
   - Resource cleanup

2. **Message Handling**
   - Buffer overflow protection
   - Failed delivery handling
   - Client removal on errors

## Usage Examples

### Client Connection

```javascript
const ws = new WebSocket("ws://api.nivai.com/ws");

ws.onopen = () => {
  console.log("Connected to WebSocket");
};

ws.onmessage = (event) => {
  console.log("Received:", event.data);
};

ws.onclose = () => {
  console.log("Disconnected from WebSocket");
};
```

### Server Broadcasting

```go
// Broadcast to all clients
hub.broadcast <- []byte("Update notification")
```

## Security Considerations

1. **Connection Security**

   - TLS encryption required
   - Origin validation needed
   - Authentication integration

2. **Message Validation**
   - Input sanitization
   - Size limits
   - Rate limiting

## Related Files

- `routes/routes.go`: WebSocket route registration
- `middleware/middleware.go`: WebSocket middleware
- `services/video_service.go`: Real-time video updates
