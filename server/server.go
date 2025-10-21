package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
	
	"go-showcase/middleware"
	ws "go-showcase/websocket"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	mu     sync.RWMutex
	users  map[int]User
	nextID int
	stats  Stats
}

type Stats struct {
	TotalRequests int       `json:"total_requests"`
	StartTime     time.Time `json:"start_time"`
	Uptime        string    `json:"uptime"`
}

var store = &Store{
	users:  make(map[int]User),
	nextID: 1,
	stats: Stats{
		StartTime: time.Now(),
	},
}

var (
	hub      *ws.Hub
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func StartServer() {
	hub = ws.NewHub()
	go hub.Run()
	
	router := mux.NewRouter()
	
	rateLimiter := middleware.NewRateLimiter(rate.Limit(10), 20)
	rateLimiter.CleanupOldVisitors()
	logger := middleware.NewRequestLogger()
	
	router.Use(middleware.Recovery)
	router.Use(middleware.CORS)
	router.Use(middleware.SecurityHeaders)
	router.Use(logger.Middleware)
	router.Use(rateLimiter.Middleware)
	
	router.HandleFunc("/api/users", getUsers).Methods("GET")
	router.HandleFunc("/api/users", createUser).Methods("POST")
	router.HandleFunc("/api/users/{id}", getUser).Methods("GET")
	router.HandleFunc("/api/users/{id}", updateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id}", deleteUser).Methods("DELETE")
	router.HandleFunc("/api/stats", getStats).Methods("GET")
	router.HandleFunc("/ws", handleWebSocket)
	router.HandleFunc("/", homeHandler).Methods("GET")
	
	initTestData()
	
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	go func() {
		fmt.Printf("🚀 Сервер запущен на http://localhost%s\n", srv.Addr)
		fmt.Println("📡 WebSocket доступен на ws://localhost:8080/ws")
		fmt.Println("⚡ Rate limiting: 10 req/s, burst: 20")
		fmt.Println("🛡️ Security headers включены")
		fmt.Println("🔄 CORS включен")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	fmt.Println("\n🛑 Graceful shutdown...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	}
	
	fmt.Println("✅ Сервер остановлен")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	client := &ws.Client{
		ID:   fmt.Sprintf("client_%d", time.Now().UnixNano()),
		Conn: conn,
		Send: make(chan ws.Message, 256),
	}
	
	hub.Register(client)
	
	go client.WritePump()
	go client.ReadPump(hub)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Showcase - Advanced Features</title>
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;600;700&family=JetBrains+Mono&display=swap" rel="stylesheet">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        :root {
            --primary: #000000;
            --secondary: #1a1a1a;
            --accent: #333333;
            --success: #666666;
            --danger: #808080;
            --warning: #999999;
            --dark: #0a0a0a;
            --light: #f5f5f5;
            --border: #2a2a2a;
        }
        
        body {
            font-family: 'Poppins', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0a0a0a;
            min-height: 100vh;
            padding: 20px;
            position: relative;
            overflow-x: hidden;
        }
        
        .particles {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            pointer-events: none;
            z-index: 1;
        }
        
        .particle {
            position: absolute;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 50%;
            animation: float 20s infinite;
        }
        
        @keyframes float {
            0%, 100% { transform: translateY(0) translateX(0); }
            25% { transform: translateY(-100px) translateX(50px); }
            50% { transform: translateY(-200px) translateX(-50px); }
            75% { transform: translateY(-100px) translateX(100px); }
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: #1a1a1a;
            border-radius: 30px;
            padding: 50px;
            box-shadow: 0 30px 80px rgba(0,0,0,0.8), 0 0 0 1px rgba(255,255,255,0.05);
            position: relative;
            z-index: 10;
            animation: slideUp 0.8s ease-out;
        }
        
        @keyframes slideUp {
            from {
                opacity: 0;
                transform: translateY(50px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        .header {
            text-align: center;
            margin-bottom: 50px;
            position: relative;
        }
        
        .logo {
            font-size: 100px;
            margin-bottom: 20px;
            animation: float3d 3s ease-in-out infinite;
            filter: drop-shadow(0 10px 20px rgba(255, 255, 255, 0.1));
            cursor: pointer;
            transition: all 0.3s ease;
        }
        
        .logo:hover {
            transform: scale(1.1) rotate(5deg);
            filter: drop-shadow(0 15px 30px rgba(255, 255, 255, 0.2));
        }
        
        @keyframes float3d {
            0%, 100% {
                transform: translateY(0) rotateZ(0deg);
            }
            25% {
                transform: translateY(-20px) rotateZ(5deg);
            }
            50% {
                transform: translateY(0) rotateZ(0deg);
            }
            75% {
                transform: translateY(-10px) rotateZ(-5deg);
            }
        }
        h1 {
            color: #ffffff;
            font-size: 3.5em;
            font-weight: 700;
            margin-bottom: 15px;
            letter-spacing: -1px;
            text-shadow: 0 0 30px rgba(255, 255, 255, 0.1);
        }
        
        .subtitle {
            color: #999999;
            font-size: 1.3em;
            font-weight: 300;
            letter-spacing: 0.5px;
        }
        h2 {
            color: #ffffff;
            margin: 40px 0 25px 0;
            padding-bottom: 15px;
            border-bottom: 2px solid #333333;
            font-size: 2em;
            font-weight: 600;
            position: relative;
            animation: slideInLeft 0.6s ease-out;
        }
        
        @keyframes slideInLeft {
            from {
                opacity: 0;
                transform: translateX(-30px);
            }
            to {
                opacity: 1;
                transform: translateX(0);
            }
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 25px;
            margin: 30px 0;
        }
        
        .feature-card {
            background: #000000;
            padding: 35px;
            border-radius: 20px;
            border: 1px solid #2a2a2a;
            color: white;
            box-shadow: 0 15px 35px rgba(0, 0, 0, 0.5);
            transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
            position: relative;
            overflow: hidden;
            cursor: pointer;
        }
        
        .feature-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(255,255,255,0.03);
            opacity: 0;
            transition: opacity 0.3s;
        }
        
        .feature-card:hover {
            transform: translateY(-10px) scale(1.02);
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.8);
            border-color: #404040;
        }
        
        .feature-card:hover::before {
            opacity: 1;
        }
        
        .feature-card h3 {
            margin-bottom: 15px;
            font-size: 1.7em;
            font-weight: 600;
        }
        
        .feature-card p {
            font-weight: 300;
            line-height: 1.6;
        }
        .endpoint {
            background: #0a0a0a;
            padding: 25px;
            margin: 20px 0;
            border-radius: 15px;
            border: 1px solid #2a2a2a;
            border-left: 4px solid #ffffff;
            transition: all 0.3s ease;
            position: relative;
            overflow: hidden;
        }
        
        .endpoint::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 4px;
            height: 100%;
            background: #ffffff;
            transition: width 0.3s ease;
        }
        
        .endpoint:hover {
            transform: translateX(5px);
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
            border-color: #404040;
        }
        
        .endpoint:hover::before {
            width: 100%;
            opacity: 0.02;
        }
        .method {
            display: inline-block;
            padding: 8px 20px;
            border-radius: 25px;
            font-weight: 700;
            margin-right: 15px;
            color: white;
            font-size: 0.95em;
            font-family: 'JetBrains Mono', monospace;
            box-shadow: 0 4px 15px rgba(0,0,0,0.2);
            transition: all 0.3s ease;
            letter-spacing: 0.5px;
        }
        
        .method:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(0,0,0,0.3);
        }
        
        .get {
            background: #ffffff;
            color: #000000;
        }
        .post {
            background: #000000;
            color: #ffffff;
            border: 1px solid #404040;
        }
        .put {
            background: #333333;
            color: #ffffff;
        }
        .delete {
            background: #666666;
            color: #ffffff;
        }
        .ws {
            background: #1a1a1a;
            color: #ffffff;
            border: 1px solid #404040;
        }
        code {
            background: #000000;
            color: #cccccc;
            padding: 6px 12px;
            border-radius: 8px;
            font-family: 'JetBrains Mono', 'Courier New', monospace;
            font-size: 0.95em;
            border: 1px solid #2a2a2a;
            box-shadow: 0 2px 10px rgba(0,0,0,0.3);
        }
        .example {
            background: #000000;
            color: #cccccc;
            padding: 25px;
            border-radius: 15px;
            margin: 20px 0;
            font-family: 'JetBrains Mono', 'Courier New', monospace;
            box-shadow: inset 0 2px 15px rgba(0,0,0,0.8);
            border: 1px solid #2a2a2a;
            overflow-x: auto;
            line-height: 1.6;
        }
        .ws-demo {
            background: #0a0a0a;
            padding: 30px;
            border-radius: 20px;
            margin: 25px 0;
            border: 1px solid #2a2a2a;
            box-shadow: 0 10px 30px rgba(0,0,0,0.5);
        }
        
        .ws-status {
            padding: 15px 20px;
            border-radius: 12px;
            margin: 15px 0;
            font-weight: 600;
            font-size: 1.1em;
            transition: all 0.3s ease;
            display: inline-block;
        }
        
        .connected {
            background: #ffffff;
            color: #000000;
            box-shadow: 0 5px 15px rgba(255, 255, 255, 0.2);
            animation: pulse 2s infinite;
        }
        
        .disconnected {
            background: #333333;
            color: #ffffff;
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.4);
        }
        #messages {
            max-height: 400px;
            overflow-y: auto;
            background: #000000;
            padding: 20px;
            border-radius: 15px;
            margin: 15px 0;
            border: 1px solid #2a2a2a;
            box-shadow: inset 0 2px 10px rgba(0,0,0,0.5);
        }
        
        #messages::-webkit-scrollbar {
            width: 8px;
        }
        
        #messages::-webkit-scrollbar-track {
            background: #1a1a1a;
            border-radius: 10px;
        }
        
        #messages::-webkit-scrollbar-thumb {
            background: #404040;
            border-radius: 10px;
        }
        
        .message {
            padding: 12px 16px;
            margin: 8px 0;
            border-radius: 12px;
            background: #1a1a1a;
            border-left: 3px solid #ffffff;
            animation: slideInRight 0.3s ease-out;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.9em;
            color: #cccccc;
        }
        
        @keyframes slideInRight {
            from {
                opacity: 0;
                transform: translateX(20px);
            }
            to {
                opacity: 1;
                transform: translateX(0);
            }
        }
        button {
            background: #ffffff;
            color: #000000;
            border: 1px solid #2a2a2a;
            padding: 14px 30px;
            border-radius: 12px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            font-family: 'Poppins', sans-serif;
            margin: 8px;
            transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
            box-shadow: 0 6px 20px rgba(0, 0, 0, 0.4);
            position: relative;
            overflow: hidden;
        }
        
        button::before {
            content: '';
            position: absolute;
            top: 50%;
            left: 50%;
            width: 0;
            height: 0;
            border-radius: 50%;
            background: rgba(0, 0, 0, 0.1);
            transform: translate(-50%, -50%);
            transition: width 0.6s, height 0.6s;
        }
        
        button:hover::before {
            width: 300px;
            height: 300px;
        }
        
        button:hover {
            transform: translateY(-3px) scale(1.05);
            box-shadow: 0 10px 30px rgba(255, 255, 255, 0.1);
            background: #f5f5f5;
        }
        
        button:active {
            transform: translateY(-1px) scale(1.02);
        }
        input {
            padding: 14px 20px;
            border: 1px solid #2a2a2a;
            border-radius: 12px;
            width: 350px;
            margin: 8px;
            font-family: 'Poppins', sans-serif;
            font-size: 15px;
            transition: all 0.3s ease;
            background: #0a0a0a;
            color: #ffffff;
        }
        
        input:focus {
            outline: none;
            border-color: #ffffff;
            box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.1);
            transform: translateY(-2px);
        }
        
        input::placeholder {
            color: #666666;
        }
    </style>
</head>
<body>
    <div class="particles" id="particles"></div>
    
    <div class="container">
        <div class="header">
            <div class="logo" onclick="this.style.transform='rotate(360deg) scale(1.2)'; setTimeout(() => this.style.transform='', 500)">🚀</div>
            <h1>Go Language Showcase</h1>
            <p class="subtitle">Advanced Features & Production-Ready Patterns</p>
        </div>
        
        <div class="grid">
            <div class="feature-card">
                <h3>🔌 WebSocket</h3>
                <p>Real-time двунаправленная коммуникация с клиентами</p>
            </div>
            <div class="feature-card">
                <h3>⚡ Rate Limiting</h3>
                <p>Защита от перегрузки: 10 req/s, burst 20</p>
            </div>
            <div class="feature-card">
                <h3>🛡️ Security</h3>
                <p>CORS, Security Headers, Recovery middleware</p>
            </div>
            <div class="feature-card">
                <h3>🔄 Graceful Shutdown</h3>
                <p>Корректное завершение всех соединений</p>
            </div>
        </div>

        <h2>🔌 WebSocket Live Demo</h2>
        <div class="ws-demo">
            <div id="status" class="ws-status disconnected">❌ Отключено</div>
            <button onclick="connectWS()">Подключиться</button>
            <button onclick="disconnectWS()">Отключиться</button>
            <button onclick="sendMessage()">Отправить тестовое сообщение</button>
            <br>
            <input type="text" id="messageInput" placeholder="Введите сообщение..." onkeypress="if(event.key==='Enter')sendCustomMessage()">
            <button onclick="sendCustomMessage()">Отправить</button>
            <div id="messages"></div>
        </div>
        
        <h2>📋 REST API Endpoints</h2>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/users</code>
            <p style="margin-top: 10px;">Получить список всех пользователей</p>
        </div>
        
        <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/users</code>
            <p style="margin-top: 10px;">Создать нового пользователя</p>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/stats</code>
            <p style="margin-top: 10px;">Статистика сервера</p>
        </div>
        
        <div class="endpoint">
            <span class="method ws">WS</span>
            <code>/ws</code>
            <p style="margin-top: 10px;">WebSocket endpoint для real-time коммуникации</p>
        </div>
        
        <h2>💡 Новые возможности</h2>
        <ul style="line-height: 2; margin: 20px; font-size: 1.1em;">
            <li>✅ <strong>WebSocket</strong> - двунаправленная real-time коммуникация</li>
            <li>✅ <strong>Rate Limiting</strong> - защита от перегрузки API</li>
            <li>✅ <strong>CORS</strong> - кросс-доменные запросы</li>
            <li>✅ <strong>Security Headers</strong> - X-Frame-Options, CSP, HSTS</li>
            <li>✅ <strong>Graceful Shutdown</strong> - корректное завершение</li>
            <li>✅ <strong>Structured Logging</strong> - детальное логирование запросов</li>
            <li>✅ <strong>Recovery Middleware</strong> - обработка паники</li>
            <li>✅ <strong>Продвинутые паттерны</strong> - Pipeline, Fan-Out/Fan-In, Circuit Breaker</li>
        </ul>
    </div>

    <script>
        const particlesContainer = document.getElementById('particles');
        for (let i = 0; i < 50; i++) {
            const particle = document.createElement('div');
            particle.className = 'particle';
            particle.style.width = Math.random() * 10 + 5 + 'px';
            particle.style.height = particle.style.width;
            particle.style.left = Math.random() * 100 + '%';
            particle.style.top = Math.random() * 100 + '%';
            particle.style.animationDelay = Math.random() * 20 + 's';
            particle.style.animationDuration = (Math.random() * 10 + 15) + 's';
            particlesContainer.appendChild(particle);
        }
        
        let ws = null;
        
        function connectWS() {
            ws = new WebSocket('ws://localhost:8080/ws');
            
            ws.onopen = function() {
                document.getElementById('status').className = 'ws-status connected';
                document.getElementById('status').innerHTML = '✅ Подключено';
                addMessage('Система', 'Подключено к WebSocket серверу', 'info');
            };
            
            ws.onmessage = function(event) {
                const data = JSON.parse(event.data);
                addMessage('Сервер', JSON.stringify(data, null, 2), 'server');
            };
            
            ws.onclose = function() {
                document.getElementById('status').className = 'ws-status disconnected';
                document.getElementById('status').innerHTML = '❌ Отключено';
                addMessage('Система', 'Отключено от сервера', 'info');
            };
            
            ws.onerror = function(error) {
                addMessage('Ошибка', 'WebSocket error: ' + error, 'error');
            };
        }
        
        function disconnectWS() {
            if (ws) {
                ws.close();
                ws = null;
            }
        }
        
        function sendMessage() {
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                alert('Сначала подключитесь к WebSocket!');
                return;
            }
            
            const msg = {
                type: 'test',
                data: { message: 'Тестовое сообщение от клиента', timestamp: new Date().toISOString() }
            };
            ws.send(JSON.stringify(msg));
            addMessage('Вы', JSON.stringify(msg, null, 2), 'client');
        }
        
        function sendCustomMessage() {
            const input = document.getElementById('messageInput');
            if (!input.value) return;
            
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                alert('Сначала подключитесь к WebSocket!');
                return;
            }
            
            const msg = {
                type: 'custom',
                data: { text: input.value }
            };
            ws.send(JSON.stringify(msg));
            addMessage('Вы', input.value, 'client');
            input.value = '';
        }
        
        function addMessage(sender, message, type) {
            const messagesDiv = document.getElementById('messages');
            const messageEl = document.createElement('div');
            messageEl.className = 'message';
            messageEl.innerHTML = '<strong>' + sender + ':</strong> ' + message;
            messagesDiv.appendChild(messageEl);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	
	users := make([]User, 0, len(store.users))
	for _, user := range store.users {
		users = append(users, user)
	}
	
	respondJSON(w, http.StatusOK, users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Неверный ID")
		return
	}
	
	store.mu.RLock()
	user, exists := store.users[id]
	store.mu.RUnlock()
	
	if !exists {
		respondError(w, http.StatusNotFound, "Пользователь не найден")
		return
	}
	
	respondJSON(w, http.StatusOK, user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}
	
	if input.Name == "" || input.Email == "" {
		respondError(w, http.StatusBadRequest, "Имя и email обязательны")
		return
	}
	
	store.mu.Lock()
	user := User{
		ID:        store.nextID,
		Name:      input.Name,
		Email:     input.Email,
		CreatedAt: time.Now(),
	}
	store.users[store.nextID] = user
	store.nextID++
	store.mu.Unlock()
	
	if hub != nil {
		hub.BroadcastMessage(ws.Message{
			Type: "user_created",
			Data: user,
			Timestamp: time.Now(),
		})
	}
	
	respondJSON(w, http.StatusCreated, user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Неверный ID")
		return
	}
	
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}
	
	store.mu.Lock()
	user, exists := store.users[id]
	if !exists {
		store.mu.Unlock()
		respondError(w, http.StatusNotFound, "Пользователь не найден")
		return
	}
	
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	store.users[id] = user
	store.mu.Unlock()
	
	respondJSON(w, http.StatusOK, user)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Неверный ID")
		return
	}
	
	store.mu.Lock()
	_, exists := store.users[id]
	if !exists {
		store.mu.Unlock()
		respondError(w, http.StatusNotFound, "Пользователь не найден")
		return
	}
	
	delete(store.users, id)
	store.mu.Unlock()
	
	respondJSON(w, http.StatusOK, map[string]string{"message": "Пользователь удален"})
}

func getStats(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	stats := store.stats
	stats.Uptime = time.Since(stats.StartTime).Round(time.Second).String()
	store.mu.RUnlock()
	
	wsStats := map[string]interface{}{}
	if hub != nil {
		wsStats = hub.GetStats()
	}
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"http":      stats,
		"websocket": wsStats,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func initTestData() {
	store.users[1] = User{
		ID:        1,
		Name:      "Иван Петров",
		Email:     "ivan@example.com",
		CreatedAt: time.Now(),
	}
	store.users[2] = User{
		ID:        2,
		Name:      "Мария Сидорова",
		Email:     "maria@example.com",
		CreatedAt: time.Now(),
	}
	store.users[3] = User{
		ID:        3,
		Name:      "Петр Иванов",
		Email:     "petr@example.com",
		CreatedAt: time.Now(),
	}
	store.nextID = 4
}
