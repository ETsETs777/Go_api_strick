# 📚 API Documentation

## Base URL
```
http://localhost:8080
```

---

## 📋 Table of Contents
- [User Management](#user-management)
- [Batch Operations](#batch-operations)
- [Search & Filter](#search--filter)
- [User Activation](#user-activation)
- [System Info](#system-info)

---

## 🔐 User Management

### Get All Users (Paginated)
```http
GET /api/users?page=1&per_page=10
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 10, max: 100)

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Иван Петров",
      "email": "ivan@example.com",
      "age": 30,
      "country": "Russia",
      "active": true,
      "created_at": "2025-10-21T19:32:00Z",
      "updated_at": "2025-10-21T19:32:00Z"
    }
  ],
  "page": 1,
  "per_page": 10,
  "total": 5,
  "total_pages": 1
}
```

---

### Get Single User
```http
GET /api/users/{id}
```

**Response:**
```json
{
  "id": 1,
  "name": "Иван Петров",
  "email": "ivan@example.com",
  "age": 30,
  "country": "Russia",
  "active": true,
  "created_at": "2025-10-21T19:32:00Z",
  "updated_at": "2025-10-21T19:32:00Z"
}
```

---

### Create User
```http
POST /api/users
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Test User",
  "email": "test@example.com",
  "age": 25,
  "country": "USA"
}
```

**Validation:**
- `name` (required): 2-100 characters
- `email` (required): Valid email format
- `age` (optional): 0-150
- `country` (optional): Any string

**Response:**
```json
{
  "id": 6,
  "name": "Test User",
  "email": "test@example.com",
  "age": 25,
  "country": "USA",
  "active": true,
  "created_at": "2025-10-21T19:35:00Z",
  "updated_at": "2025-10-21T19:35:00Z"
}
```

---

### Update User
```http
PUT /api/users/{id}
Content-Type: application/json
```

**Request Body (all fields optional):**
```json
{
  "name": "Updated Name",
  "email": "updated@example.com",
  "age": 26,
  "country": "Canada"
}
```

**Response:**
```json
{
  "id": 1,
  "name": "Updated Name",
  "email": "updated@example.com",
  "age": 26,
  "country": "Canada",
  "active": true,
  "created_at": "2025-10-21T19:32:00Z",
  "updated_at": "2025-10-21T19:36:00Z"
}
```

---

### Delete User
```http
DELETE /api/users/{id}
```

**Response:**
```json
{
  "message": "Пользователь удален"
}
```

---

## 🔄 Batch Operations

### Batch Create Users
```http
POST /api/users/batch
Content-Type: application/json
```

**Request Body:**
```json
{
  "users": [
    {
      "name": "User 1",
      "email": "user1@example.com",
      "age": 25,
      "country": "USA"
    },
    {
      "name": "User 2",
      "email": "user2@example.com",
      "age": 30,
      "country": "UK"
    }
  ]
}
```

**Limits:**
- Maximum 100 users per batch
- Invalid users are skipped

**Response:**
```json
{
  "created": [
    {
      "id": 6,
      "name": "User 1",
      "email": "user1@example.com",
      "age": 25,
      "country": "USA",
      "active": true,
      "created_at": "2025-10-21T19:40:00Z",
      "updated_at": "2025-10-21T19:40:00Z"
    },
    {
      "id": 7,
      "name": "User 2",
      "email": "user2@example.com",
      "age": 30,
      "country": "UK",
      "active": true,
      "created_at": "2025-10-21T19:40:00Z",
      "updated_at": "2025-10-21T19:40:00Z"
    }
  ],
  "count": 2
}
```

---

### Batch Delete Users
```http
DELETE /api/users/batch
Content-Type: application/json
```

**Request Body:**
```json
{
  "ids": [1, 2, 3, 4]
}
```

**Response:**
```json
{
  "deleted": [1, 2, 3, 4],
  "count": 4
}
```

---

## 🔍 Search & Filter

### Search Users
```http
GET /api/users/search?q=john&country=USA&active=true
```

**Query Parameters:**
- `q` (optional): Search in name and email
- `country` (optional): Filter by country
- `active` (optional): Filter by active status (true/false)

**Response:**
```json
{
  "results": [
    {
      "id": 4,
      "name": "John Smith",
      "email": "john@example.com",
      "age": 28,
      "country": "USA",
      "active": true,
      "created_at": "2025-10-21T19:32:00Z",
      "updated_at": "2025-10-21T19:32:00Z"
    }
  ],
  "count": 1
}
```

---

## ⚡ User Activation

### Activate User
```http
PATCH /api/users/{id}/activate
```

**Response:**
```json
{
  "id": 3,
  "name": "Петр Иванов",
  "email": "petr@example.com",
  "age": 35,
  "country": "Ukraine",
  "active": true,
  "created_at": "2025-10-21T19:32:00Z",
  "updated_at": "2025-10-21T19:45:00Z"
}
```

---

### Deactivate User
```http
PATCH /api/users/{id}/deactivate
```

**Response:**
```json
{
  "id": 1,
  "name": "Иван Петров",
  "email": "ivan@example.com",
  "age": 30,
  "country": "Russia",
  "active": false,
  "created_at": "2025-10-21T19:32:00Z",
  "updated_at": "2025-10-21T19:46:00Z"
}
```

---

## 📊 System Info

### Health Check
```http
GET /api/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-21T19:50:00Z",
  "uptime": "18m30s",
  "total_users": 5,
  "active_users": 4
}
```

---

### Server Statistics
```http
GET /api/stats
```

**Response:**
```json
{
  "http": {
    "total_requests": 150,
    "total_users": 5,
    "active_users": 4,
    "users_by_country": {
      "Russia": 2,
      "Ukraine": 1,
      "USA": 1,
      "Germany": 1
    },
    "start_time": "2025-10-21T19:32:00Z",
    "uptime": "18m45s"
  },
  "websocket": {
    "total_clients": 2,
    "timestamp": "2025-10-21T19:50:45Z"
  }
}
```

---

## 🌐 WebSocket

### Connect to WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

// Send message
ws.send(JSON.stringify({
  type: 'custom',
  data: { text: 'Hello, Server!' }
}));
```

**Message Types:**
- `welcome` - Welcome message on connect
- `user_created` - Broadcast when user is created
- `heartbeat` - Periodic server heartbeat (every 30s)
- `echo` - Echo back custom messages
- `shutdown` - Server shutdown notification

---

## ⚠️ Error Responses

### Bad Request (400)
```json
{
  "error": "Invalid JSON format"
}
```

### Not Found (404)
```json
{
  "error": "User not found"
}
```

### Too Many Requests (429)
```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

**Headers:**
- `Retry-After: 60` (seconds)

### Internal Server Error (500)
```json
{
  "error": "Internal server error occurred"
}
```

---

## 🔧 Rate Limiting

- **Rate**: 10 requests per second
- **Burst**: 20 requests
- Applied to all API endpoints

---

## 📝 Examples (cURL)

### Create User
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Johnson",
    "email": "alice@example.com",
    "age": 28,
    "country": "Canada"
  }'
```

### Search Users
```bash
curl "http://localhost:8080/api/users/search?q=john&active=true"
```

### Batch Create
```bash
curl -X POST http://localhost:8080/api/users/batch \
  -H "Content-Type: application/json" \
  -d '{
    "users": [
      {"name": "User 1", "email": "user1@test.com", "age": 25},
      {"name": "User 2", "email": "user2@test.com", "age": 30}
    ]
  }'
```

### Get Paginated Users
```bash
curl "http://localhost:8080/api/users?page=1&per_page=5"
```

### Batch Delete
```bash
curl -X DELETE http://localhost:8080/api/users/batch \
  -H "Content-Type: application/json" \
  -d '{"ids": [1, 2, 3]}'
```

---

## 🎨 Testing via Web Interface

Visit `http://localhost:8080` for the interactive web interface where you can:
- View API documentation
- Test WebSocket connection
- See real-time server statistics
- Test all endpoints visually

---

**Last Updated**: October 21, 2025  
**API Version**: 2.0  
**Status**: ✅ Production Ready

