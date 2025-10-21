# 🚀 Go Language Showcase

A comprehensive project demonstrating all major features of the Go programming language.

## 📋 Table of Contents

This project includes examples and demonstrations of the following Go features:

### 1. **Data Types** (`types/`)
- Basic types (int, float, string, bool, complex)
- Structs
- Arrays & slices
- Maps
- Constants and variables

### 2. **Interfaces** (`interfaces/`)
- Interface definition and implementation
- Polymorphism
- Type assertions and type switches
- Error handling (error interface)
- Custom errors

### 3. **Concurrency** (`concurrency/`)
- Goroutines
- Channels
- Buffered and unbuffered channels
- Select statement
- Worker Pool pattern
- Mutexes and synchronization (sync.Mutex, sync.WaitGroup)

### 4. **Generics** (`generics/`)
- Generic functions
- Generic structs (Stack, Map)
- Type constraints
- Parameterized types

### 5. **Reflection** (`reflection/`)
- Getting types and values
- Struct inspection
- Working with struct tags
- Modifying values through reflection
- Calling methods dynamically

### 6. **Database** (`database/`)
- Working with SQLite
- CRUD operations
- Transactions
- Prepared statements
- JSON data export

### 7. **HTTP Server** (`server/`)
- REST API with Gorilla Mux
- WebSocket for real-time communication
- Rate Limiting (10 req/s)
- CORS middleware
- Security Headers
- Graceful Shutdown
- Structured Logging
- JSON encoding/decoding
- Beautiful monochrome web interface

### 8. **Advanced Concurrency Patterns** (`advanced/`)
- Pipeline Pattern
- Fan-Out/Fan-In Pattern
- Circuit Breaker Pattern
- Semaphore Pattern
- In-Memory Cache with TTL

### 9. **Other Features** (`main.go`)
- Defer, panic, recover
- File operations (read, write)
- Context (with timeout and cancellation)
- Memory management

## 🔧 Installation

### Step 1: Install Go

#### Windows
1. Download Go from the official website: https://go.dev/dl/
2. Run the installer (double-click the `.msi` file)
3. Follow the installation wizard instructions
4. After installation, open a new terminal window

#### Linux
```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### macOS
```bash
brew install go
```

### Step 2: Verify Installation
```bash
go version
```

You should see something like: `go version go1.21.5 windows/amd64`

### Step 3: Install Project Dependencies
```bash
go mod download
```

## 🚀 Usage

### Run the Entire Project
```bash
go run main.go
```

This will run all demonstrations sequentially and finally start the HTTP server.

### HTTP Server
After running `go run main.go`, the HTTP server will be available at:
```
http://localhost:8080
```

Open your browser and navigate to this address to see the beautiful web interface with API documentation.

### API Endpoints

- `GET /api/users` - Get all users
- `GET /api/users/{id}` - Get user by ID
- `POST /api/users` - Create a new user
- `PUT /api/users/{id}` - Update a user
- `DELETE /api/users/{id}` - Delete a user
- `GET /api/stats` - Get server statistics
- `WS /ws` - WebSocket connection for real-time communication

### API Usage Examples

#### Get All Users
```bash
curl http://localhost:8080/api/users
```

#### Create a User
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Anna Ivanova","email":"anna@example.com"}'
```

#### Get User by ID
```bash
curl http://localhost:8080/api/users/1
```

#### Update a User
```bash
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"New Name","email":"new@example.com"}'
```

#### Delete a User
```bash
curl -X DELETE http://localhost:8080/api/users/1
```

#### Get Statistics
```bash
curl http://localhost:8080/api/stats
```

## 📦 Project Structure

```
go-showcase/
├── main.go                 # Main application file
├── go.mod                  # Dependencies file
├── go.sum                  # Dependencies checksums
├── README.md               # This file
├── .gitignore              # Git ignore file
├── types/                  # Data types demonstration
│   └── basic_types.go
├── interfaces/             # Interfaces and error handling
│   └── interfaces.go
├── concurrency/            # Goroutines and channels
│   └── concurrency.go
├── generics/               # Generics (Go 1.18+)
│   └── generics.go
├── reflection/             # Reflection
│   └── reflection.go
├── database/               # Database operations
│   └── database.go
├── server/                 # HTTP server
│   └── server.go
├── advanced/               # Advanced patterns
│   └── patterns.go
├── middleware/             # HTTP middleware
│   └── middleware.go
└── websocket/              # WebSocket hub
    └── hub.go
```

## 🎯 What This Project Demonstrates

### Basic Concepts
- ✅ Data types and variables
- ✅ Structs and methods
- ✅ Interfaces and polymorphism
- ✅ Arrays, slices, and maps
- ✅ Pointers

### Advanced Features
- ✅ Goroutines and channels
- ✅ Select statement
- ✅ Mutexes and synchronization
- ✅ Generics (Go 1.18+)
- ✅ Reflection
- ✅ Context

### Practical Applications
- ✅ HTTP server with REST API
- ✅ Database operations (SQLite)
- ✅ JSON encoding/decoding
- ✅ File operations
- ✅ Middleware
- ✅ Error handling

### Patterns and Best Practices
- ✅ Worker Pool
- ✅ Dependency Injection
- ✅ Error wrapping
- ✅ Defer, panic, recover
- ✅ Package organization

## 📚 Additional Resources

- [Official Go Documentation](https://go.dev/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Tour](https://go.dev/tour/)

## 🤝 Contributing

This project was created for educational purposes to demonstrate Go capabilities.

## 📝 License

MIT License - feel free to use this code for learning and practice!

## ⚡ Quick Start

1. Install Go (download from go.dev)
2. Clone or download this project
3. Open terminal in the project folder
4. Run: `go mod download`
5. Start: `go run main.go`
6. Open browser: `http://localhost:8080`

Enjoy learning Go! 🎉

