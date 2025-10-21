package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
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
	Age       int       `json:"age,omitempty"`
	Country   string    `json:"country,omitempty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaginatedResponse struct {
	Data       []User `json:"data"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	Total      int    `json:"total"`
	TotalPages int    `json:"total_pages"`
}

type BatchCreateRequest struct {
	Users []struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Age     int    `json:"age,omitempty"`
		Country string `json:"country,omitempty"`
	} `json:"users"`
}

type BatchDeleteRequest struct {
	IDs []int `json:"ids"`
}

type Store struct {
	mu     sync.RWMutex
	users  map[int]User
	nextID int
	stats  Stats
}

type Stats struct {
	TotalRequests   int            `json:"total_requests"`
	TotalUsers      int            `json:"total_users"`
	ActiveUsers     int            `json:"active_users"`
	UsersByCountry  map[string]int `json:"users_by_country"`
	RequestsByPath  map[string]int `json:"requests_by_path"`
	AvgResponseTime float64        `json:"avg_response_time_ms"`
	StartTime       time.Time      `json:"start_time"`
	Uptime          string         `json:"uptime"`
}

type PerformanceMetrics struct {
	Path         string
	Count        int
	TotalTime    time.Duration
	AverageTime  time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
}

var (
	metrics      = make(map[string]*PerformanceMetrics)
	metricsMutex sync.RWMutex
)

var store = &Store{
	users:  make(map[int]User),
	nextID: 1,
	stats: Stats{
		StartTime:      time.Now(),
		RequestsByPath: make(map[string]int),
		UsersByCountry: make(map[string]int),
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
	router.HandleFunc("/api/users/batch", batchCreateUsers).Methods("POST")
	router.HandleFunc("/api/users/batch", batchDeleteUsers).Methods("DELETE")
	router.HandleFunc("/api/users/search", searchUsers).Methods("GET")
	router.HandleFunc("/api/users/export", exportUsers).Methods("GET")
	router.HandleFunc("/api/users/analytics", getUserAnalytics).Methods("GET")
	router.HandleFunc("/api/users/{id}", getUser).Methods("GET")
	router.HandleFunc("/api/users/{id}", updateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id}/activate", activateUser).Methods("PATCH")
	router.HandleFunc("/api/users/{id}/deactivate", deactivateUser).Methods("PATCH")
	router.HandleFunc("/api/users/{id}", deleteUser).Methods("DELETE")
	router.HandleFunc("/api/stats", getStats).Methods("GET")
	router.HandleFunc("/api/metrics", getMetrics).Methods("GET")
	router.HandleFunc("/api/health", healthCheck).Methods("GET")
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
	
	fmt.Println("\n🛑 Graceful shutdown initiated...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	fmt.Println("   Closing WebSocket connections...")
	hub.Shutdown()
	
	fmt.Println("   Stopping HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Server shutdown error: %v", err)
	} else {
		fmt.Println("✅ Server stopped gracefully")
	}
	
	fmt.Println("👋 Goodbye!")
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
            --shadow-sm: 0 2px 8px rgba(0, 0, 0, 0.4);
            --shadow-md: 0 8px 24px rgba(0, 0, 0, 0.6);
            --shadow-lg: 0 16px 48px rgba(0, 0, 0, 0.8);
            --shadow-xl: 0 24px 64px rgba(0, 0, 0, 0.9);
        }
        
        body {
            font-family: 'Poppins', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #000000;
            background-image: 
                radial-gradient(circle at 20% 50%, rgba(255, 255, 255, 0.03) 0%, transparent 50%),
                radial-gradient(circle at 80% 80%, rgba(255, 255, 255, 0.02) 0%, transparent 50%);
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
            background: rgba(255, 255, 255, 0.08);
            border-radius: 50%;
            animation: float 20s infinite ease-in-out;
            box-shadow: 0 0 20px rgba(255, 255, 255, 0.1);
        }
        
        @keyframes float {
            0%, 100% { 
                transform: translateY(0) translateX(0) scale(1); 
                opacity: 0.3;
            }
            25% { 
                transform: translateY(-100px) translateX(50px) scale(1.2); 
                opacity: 0.6;
            }
            50% { 
                transform: translateY(-200px) translateX(-50px) scale(0.8); 
                opacity: 0.4;
            }
            75% { 
                transform: translateY(-100px) translateX(100px) scale(1.1); 
                opacity: 0.5;
            }
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: linear-gradient(135deg, #1a1a1a 0%, #0f0f0f 100%);
            border-radius: 40px;
            padding: 60px;
            box-shadow: 
                0 50px 100px rgba(0,0,0,0.9),
                0 0 0 1px rgba(255,255,255,0.08),
                inset 0 1px 0 rgba(255,255,255,0.05);
            position: relative;
            z-index: 10;
            animation: slideUp 0.8s cubic-bezier(0.16, 1, 0.3, 1);
            backdrop-filter: blur(10px);
        }
        
        .container::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            border-radius: 40px;
            padding: 2px;
            background: linear-gradient(135deg, rgba(255,255,255,0.1), rgba(255,255,255,0.02));
            -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
            -webkit-mask-composite: xor;
            mask-composite: exclude;
            pointer-events: none;
        }
        
        @keyframes slideUp {
            from {
                opacity: 0;
                transform: translateY(60px) scale(0.95);
                filter: blur(10px);
            }
            to {
                opacity: 1;
                transform: translateY(0) scale(1);
                filter: blur(0);
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
            letter-spacing: -2px;
            text-shadow: 
                0 0 40px rgba(255, 255, 255, 0.15),
                0 2px 4px rgba(0, 0, 0, 0.5);
            background: linear-gradient(135deg, #ffffff 0%, #cccccc 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .subtitle {
            color: #999999;
            font-size: 1.3em;
            font-weight: 300;
            letter-spacing: 0.5px;
        }
        h2 {
            color: #ffffff;
            margin: 50px 0 30px 0;
            padding-bottom: 20px;
            border-bottom: 2px solid transparent;
            background: linear-gradient(90deg, #333333 0%, transparent 100%) bottom / 100% 2px no-repeat;
            font-size: 2em;
            font-weight: 600;
            position: relative;
            animation: slideInLeft 0.6s cubic-bezier(0.16, 1, 0.3, 1);
            letter-spacing: -0.5px;
        }
        
        h2::before {
            content: '';
            position: absolute;
            left: 0;
            bottom: -2px;
            width: 80px;
            height: 2px;
            background: linear-gradient(90deg, #ffffff 0%, transparent 100%);
            animation: slideWidth 1s ease-out;
        }
        
        @keyframes slideWidth {
            from { width: 0; }
            to { width: 80px; }
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
            background: linear-gradient(135deg, #0a0a0a 0%, #000000 100%);
            padding: 40px;
            border-radius: 24px;
            border: 1px solid #2a2a2a;
            color: white;
            box-shadow: var(--shadow-md);
            transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
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
            background: radial-gradient(circle at var(--mouse-x, 50%) var(--mouse-y, 50%), rgba(255,255,255,0.08) 0%, transparent 50%);
            opacity: 0;
            transition: opacity 0.5s;
        }
        
        .feature-card::after {
            content: '';
            position: absolute;
            top: -50%;
            left: -50%;
            width: 200%;
            height: 200%;
            background: conic-gradient(from 0deg at 50% 50%, transparent 0deg, rgba(255,255,255,0.05) 180deg, transparent 360deg);
            animation: rotate 8s linear infinite;
            opacity: 0;
            transition: opacity 0.5s;
        }
        
        @keyframes rotate {
            100% { transform: rotate(360deg); }
        }
        
        .feature-card:hover {
            transform: translateY(-12px) scale(1.03);
            box-shadow: 
                var(--shadow-xl),
                0 0 0 1px rgba(255,255,255,0.1);
            border-color: #505050;
        }
        
        .feature-card:hover::before {
            opacity: 1;
        }
        
        .feature-card:hover::after {
            opacity: 1;
        }
        
        .feature-card h3 {
            margin-bottom: 18px;
            font-size: 1.8em;
            font-weight: 600;
            letter-spacing: -0.5px;
            position: relative;
            z-index: 1;
        }
        
        .feature-card p {
            font-weight: 300;
            line-height: 1.7;
            color: #cccccc;
            position: relative;
            z-index: 1;
        }
        .endpoint {
            background: linear-gradient(135deg, #0a0a0a 0%, #050505 100%);
            padding: 28px;
            margin: 20px 0;
            border-radius: 18px;
            border: 1px solid #2a2a2a;
            border-left: 4px solid #ffffff;
            transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
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
            background: linear-gradient(180deg, #ffffff 0%, #999999 100%);
            transition: all 0.4s ease;
            box-shadow: 0 0 20px rgba(255,255,255,0.3);
        }
        
        .endpoint:hover {
            transform: translateX(8px);
            box-shadow: 
                var(--shadow-lg),
                -4px 0 20px rgba(255,255,255,0.05);
            border-color: #505050;
        }
        
        .endpoint:hover::before {
            width: 100%;
            opacity: 0.03;
        }
        .method {
            display: inline-block;
            padding: 10px 24px;
            border-radius: 30px;
            font-weight: 700;
            margin-right: 15px;
            color: white;
            font-size: 0.9em;
            font-family: 'JetBrains Mono', monospace;
            box-shadow: var(--shadow-sm);
            transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
            letter-spacing: 1px;
            text-transform: uppercase;
            position: relative;
            overflow: hidden;
        }
        
        .method::before {
            content: '';
            position: absolute;
            top: 50%;
            left: 50%;
            width: 0;
            height: 0;
            border-radius: 50%;
            background: rgba(255,255,255,0.1);
            transform: translate(-50%, -50%);
            transition: width 0.6s, height 0.6s;
        }
        
        .method:hover {
            transform: translateY(-3px) scale(1.05);
            box-shadow: var(--shadow-md);
        }
        
        .method:hover::before {
            width: 300px;
            height: 300px;
        }
        
        .get {
            background: linear-gradient(135deg, #ffffff 0%, #e0e0e0 100%);
            color: #000000;
            box-shadow: 0 4px 15px rgba(255,255,255,0.1);
        }
        
        .get:hover {
            background: linear-gradient(135deg, #e0e0e0 0%, #ffffff 100%);
        }
        
        .post {
            background: linear-gradient(135deg, #000000 0%, #1a1a1a 100%);
            color: #ffffff;
            border: 1px solid #404040;
            box-shadow: 0 4px 15px rgba(0,0,0,0.5);
        }
        
        .post:hover {
            background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
        }
        
        .put {
            background: linear-gradient(135deg, #333333 0%, #404040 100%);
            color: #ffffff;
            box-shadow: 0 4px 15px rgba(51,51,51,0.3);
        }
        
        .put:hover {
            background: linear-gradient(135deg, #404040 0%, #505050 100%);
        }
        
        .delete {
            background: linear-gradient(135deg, #666666 0%, #808080 100%);
            color: #ffffff;
            box-shadow: 0 4px 15px rgba(102,102,102,0.3);
        }
        
        .delete:hover {
            background: linear-gradient(135deg, #808080 0%, #999999 100%);
        }
        
        .ws {
            background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
            color: #ffffff;
            border: 1px solid #505050;
            box-shadow: 0 4px 15px rgba(26,26,26,0.4);
        }
        
        .ws:hover {
            background: linear-gradient(135deg, #2a2a2a 0%, #3a3a3a 100%);
        }
        code {
            background: linear-gradient(135deg, #0a0a0a 0%, #000000 100%);
            color: #e0e0e0;
            padding: 8px 14px;
            border-radius: 10px;
            font-family: 'JetBrains Mono', 'Courier New', monospace;
            font-size: 0.95em;
            border: 1px solid #2a2a2a;
            box-shadow: 
                var(--shadow-sm),
                inset 0 1px 0 rgba(255,255,255,0.03);
            letter-spacing: 0.3px;
        }
        
        .example {
            background: linear-gradient(135deg, #000000 0%, #0a0a0a 100%);
            color: #d0d0d0;
            padding: 28px;
            border-radius: 18px;
            margin: 20px 0;
            font-family: 'JetBrains Mono', 'Courier New', monospace;
            box-shadow: 
                inset 0 2px 15px rgba(0,0,0,0.9),
                0 4px 20px rgba(0,0,0,0.6);
            border: 1px solid #2a2a2a;
            overflow-x: auto;
            line-height: 1.7;
            position: relative;
        }
        
        .example::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 1px;
            background: linear-gradient(90deg, transparent, rgba(255,255,255,0.05), transparent);
        }
        .ws-demo {
            background: linear-gradient(135deg, #0a0a0a 0%, #000000 100%);
            padding: 35px;
            border-radius: 24px;
            margin: 30px 0;
            border: 1px solid #2a2a2a;
            box-shadow: var(--shadow-lg);
            position: relative;
            overflow: hidden;
        }
        
        .ws-demo::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: radial-gradient(circle at 50% 0%, rgba(255,255,255,0.03) 0%, transparent 50%);
            pointer-events: none;
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
            background: linear-gradient(135deg, #ffffff 0%, #e0e0e0 100%);
            color: #000000;
            box-shadow: 
                0 8px 24px rgba(255, 255, 255, 0.15),
                0 0 30px rgba(255, 255, 255, 0.1);
            animation: pulse 2s infinite, glow 2s ease-in-out infinite;
            font-weight: 700;
        }
        
        .disconnected {
            background: linear-gradient(135deg, #333333 0%, #2a2a2a 100%);
            color: #999999;
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.5);
        }
        
        @keyframes glow {
            0%, 100% { box-shadow: 0 8px 24px rgba(255, 255, 255, 0.15), 0 0 30px rgba(255, 255, 255, 0.1); }
            50% { box-shadow: 0 8px 24px rgba(255, 255, 255, 0.25), 0 0 40px rgba(255, 255, 255, 0.2); }
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
            background: linear-gradient(135deg, #ffffff 0%, #f0f0f0 100%);
            color: #000000;
            border: 1px solid #2a2a2a;
            padding: 14px 32px;
            border-radius: 14px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            font-family: 'Poppins', sans-serif;
            margin: 8px;
            transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
            box-shadow: var(--shadow-sm);
            position: relative;
            overflow: hidden;
            letter-spacing: 0.3px;
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
            background: linear-gradient(135deg, #f0f0f0 0%, #e0e0e0 100%);
            transform: translateY(-4px) scale(1.02);
            box-shadow: var(--shadow-md);
            border-color: #404040;
        }
        
        button:active {
            transform: translateY(-1px) scale(0.98);
            box-shadow: var(--shadow-sm);
        }
        input {
            padding: 14px 22px;
            border: 1px solid #2a2a2a;
            border-radius: 14px;
            width: 350px;
            margin: 8px;
            font-family: 'Poppins', sans-serif;
            font-size: 15px;
            transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
            background: linear-gradient(135deg, #0a0a0a 0%, #050505 100%);
            color: #ffffff;
            box-shadow: inset 0 2px 10px rgba(0,0,0,0.5);
        }
        
        input:focus {
            outline: none;
            border-color: #ffffff;
            background: linear-gradient(135deg, #0f0f0f 0%, #0a0a0a 100%);
            box-shadow: 
                0 0 0 4px rgba(255,255,255,0.08),
                inset 0 2px 10px rgba(0,0,0,0.5),
                0 8px 24px rgba(255,255,255,0.05);
            transform: translateY(-2px);
        }
        
        input::placeholder {
            color: #666666;
        }
        
        .tabs {
            display: flex;
            gap: 10px;
            margin: 30px 0;
            border-bottom: 2px solid #2a2a2a;
            flex-wrap: wrap;
        }
        
        .tab {
            padding: 14px 28px;
            background: transparent;
            color: #999999;
            border: none;
            border-bottom: 3px solid transparent;
            cursor: pointer;
            font-size: 1em;
            font-weight: 600;
            transition: all 0.3s ease;
            position: relative;
        }
        
        .tab:hover {
            color: #ffffff;
            background: linear-gradient(135deg, #1a1a1a 0%, #0f0f0f 100%);
        }
        
        .tab.active {
            color: #ffffff;
            border-bottom-color: #ffffff;
            background: linear-gradient(135deg, #2a2a2a 0%, #1a1a1a 100%);
        }
        
        .tab-content {
            display: none;
            animation: fadeIn 0.5s ease;
        }
        
        .tab-content.active {
            display: block;
        }
        
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
        
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        
        .metric-card {
            background: linear-gradient(135deg, #1a1a1a 0%, #0f0f0f 100%);
            padding: 20px;
            border-radius: 16px;
            border: 1px solid #2a2a2a;
            box-shadow: var(--shadow-md);
        }
        
        .metric-card h4 {
            color: #ffffff;
            margin-bottom: 10px;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        .metric-value {
            font-size: 2em;
            font-weight: 700;
            background: linear-gradient(135deg, #ffffff 0%, #cccccc 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .metric-label {
            color: #999999;
            font-size: 0.9em;
            margin-top: 5px;
        }
        
        .copy-btn {
            background: #2a2a2a;
            border: 1px solid #3a3a3a;
            color: #cccccc;
            padding: 6px 12px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 0.8em;
            transition: all 0.3s ease;
            margin-left: 10px;
        }
        
        .copy-btn:hover {
            background: #3a3a3a;
            color: #ffffff;
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
        
        <div class="tabs">
            <button class="tab active" onclick="showTab('overview')">📊 Overview</button>
            <button class="tab" onclick="showTab('api')">📋 API Docs</button>
            <button class="tab" onclick="showTab('tester')">🧪 Tester</button>
            <button class="tab" onclick="showTab('metrics')">⚡ Metrics</button>
            <button class="tab" onclick="showTab('websocket')">🔌 WebSocket</button>
        </div>
        
        <div id="overview" class="tab-content active">
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
            
            <h2>💡 All Features</h2>
            <ul style="line-height: 2; margin: 20px; font-size: 1.1em; columns: 2; column-gap: 40px;">
                <li>✅ <strong>Pagination</strong> - страничный вывод данных</li>
                <li>✅ <strong>Sorting</strong> - сортировка по полям</li>
                <li>✅ <strong>Search & Filter</strong> - поиск и фильтрация</li>
                <li>✅ <strong>Batch Operations</strong> - массовые операции</li>
                <li>✅ <strong>Export</strong> - экспорт в JSON/CSV</li>
                <li>✅ <strong>Analytics</strong> - детальная статистика</li>
                <li>✅ <strong>Email Validation</strong> - проверка формата</li>
                <li>✅ <strong>Age Validation</strong> - диапазон 0-150</li>
                <li>✅ <strong>WebSocket</strong> - real-time коммуникация</li>
                <li>✅ <strong>Rate Limiting</strong> - 10 req/s, burst 20</li>
                <li>✅ <strong>CORS</strong> - кросс-доменные запросы</li>
                <li>✅ <strong>Security Headers</strong> - CSP, HSTS, X-Frame</li>
                <li>✅ <strong>Graceful Shutdown</strong> - корректное завершение</li>
                <li>✅ <strong>Structured Logging</strong> - детальные логи</li>
                <li>✅ <strong>Recovery Middleware</strong> - обработка паники</li>
                <li>✅ <strong>Performance Metrics</strong> - отслеживание производительности</li>
            </ul>
        </div>
        
        <div id="api" class="tab-content">

        <h2>📋 REST API Endpoints</h2>
        
        <h3 style="color: #cccccc; margin: 30px 0 20px 0;">📄 User Management</h3>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/users?page=1&per_page=10&sort=name&order=asc</code>
            <p style="margin-top: 10px;">Get paginated users with sorting</p>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/users/{id}</code>
            <p style="margin-top: 10px;">Get single user by ID</p>
        </div>
        
        <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/users</code>
            <p style="margin-top: 10px;">Create new user (name, email, age, country)</p>
        </div>
        
        <div class="endpoint">
            <span class="method put">PUT</span>
            <code>/api/users/{id}</code>
            <p style="margin-top: 10px;">Update user information</p>
        </div>
        
        <div class="endpoint">
            <span class="method delete">DELETE</span>
            <code>/api/users/{id}</code>
            <p style="margin-top: 10px;">Delete user</p>
        </div>
        
        <h3 style="color: #cccccc; margin: 30px 0 20px 0;">🔍 Search & Filter</h3>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/users/search?q=john&country=USA&active=true</code>
            <p style="margin-top: 10px;">Search users by name, email, country, and status</p>
        </div>
        
        <h3 style="color: #cccccc; margin: 30px 0 20px 0;">🔄 Batch Operations</h3>
        
        <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/users/batch</code>
            <p style="margin-top: 10px;">Create multiple users at once (max 100)</p>
        </div>
        
        <div class="endpoint">
            <span class="method delete">DELETE</span>
            <code>/api/users/batch</code>
            <p style="margin-top: 10px;">Delete multiple users by IDs</p>
        </div>
        
        <h3 style="color: #cccccc; margin: 30px 0 20px 0;">⚡ User Actions</h3>
        
        <div class="endpoint">
            <span class="method put">PATCH</span>
            <code>/api/users/{id}/activate</code>
            <p style="margin-top: 10px;">Activate user</p>
        </div>
        
        <div class="endpoint">
            <span class="method put">PATCH</span>
            <code>/api/users/{id}/deactivate</code>
            <p style="margin-top: 10px;">Deactivate user</p>
        </div>
        
        <h3 style="color: #cccccc; margin: 30px 0 20px 0;">📊 Analytics & Export</h3>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/users/analytics</code>
            <p style="margin-top: 10px;">Get user analytics (by country, age, status)</p>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/users/export?format=csv</code>
            <p style="margin-top: 10px;">Export users as JSON or CSV</p>
        </div>
        
        <h3 style="color: #cccccc; margin: 30px 0 20px 0;">🔧 System</h3>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/stats</code>
            <p style="margin-top: 10px;">Server statistics with user breakdown</p>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/health</code>
            <p style="margin-top: 10px;">Health check endpoint</p>
        </div>
        
        <div class="endpoint">
            <span class="method ws">WS</span>
            <code>/ws</code>
            <p style="margin-top: 10px;">WebSocket real-time communication</p>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/metrics</code>
            <p style="margin-top: 10px;">Performance metrics (response times, request counts)</p>
        </div>
        </div>
        
        <div id="tester" class="tab-content">
        <h2>🧪 Interactive API Tester</h2>
        <div class="ws-demo">
            <h3 style="margin-bottom: 15px;">📥 Export Users</h3>
            <button onclick="window.open('/api/users/export?format=json', '_blank')">Export as JSON</button>
            <button onclick="window.open('/api/users/export?format=csv', '_blank')">Export as CSV</button>
            
            <h3 style="margin: 25px 0 15px 0;">📊 View Analytics</h3>
            <button onclick="fetchAnalytics()">Get Analytics</button>
            <div id="analytics" style="margin-top: 15px;"></div>
            
            <h3 style="margin: 25px 0 15px 0;">🔍 Search Users</h3>
            <input type="text" id="searchQuery" placeholder="Search by name or email...">
            <input type="text" id="searchCountry" placeholder="Filter by country...">
            <button onclick="searchUsersAPI()">Search</button>
            <div id="searchResults" style="margin-top: 15px;"></div>
        </div>
        </div>
        
        <div id="metrics" class="tab-content">
        <h2>⚡ Performance Metrics</h2>
        <div class="ws-demo">
            <button onclick="loadMetrics()">🔄 Refresh Metrics</button>
            <div id="metricsDisplay" style="margin-top: 20px;"></div>
        </div>
        </div>
        
        <div id="websocket" class="tab-content">
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
        </div>
    </div>

    <script>
        const particlesContainer = document.getElementById('particles');
        for (let i = 0; i < 60; i++) {
            const particle = document.createElement('div');
            particle.className = 'particle';
            const size = Math.random() * 6 + 2;
            particle.style.width = size + 'px';
            particle.style.height = size + 'px';
            particle.style.left = Math.random() * 100 + '%';
            particle.style.top = Math.random() * 100 + '%';
            particle.style.animationDelay = Math.random() * 20 + 's';
            particle.style.animationDuration = (Math.random() * 15 + 15) + 's';
            particle.style.opacity = Math.random() * 0.5 + 0.2;
            particlesContainer.appendChild(particle);
        }
        
        document.querySelectorAll('.feature-card').forEach(card => {
            card.addEventListener('mousemove', (e) => {
                const rect = card.getBoundingClientRect();
                const x = ((e.clientX - rect.left) / rect.width) * 100;
                const y = ((e.clientY - rect.top) / rect.height) * 100;
                card.style.setProperty('--mouse-x', x + '%');
                card.style.setProperty('--mouse-y', y + '%');
            });
        });
        
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
        
        async function fetchAnalytics() {
            try {
                const response = await fetch('/api/users/analytics');
                const data = await response.json();
                const analyticsDiv = document.getElementById('analytics');
                analyticsDiv.innerHTML = '<div class="example">' + JSON.stringify(data, null, 2) + '</div>';
            } catch (error) {
                document.getElementById('analytics').innerHTML = '<p style="color: #ff6666;">Error: ' + error.message + '</p>';
            }
        }
        
        async function searchUsersAPI() {
            const query = document.getElementById('searchQuery').value;
            const country = document.getElementById('searchCountry').value;
            const params = new URLSearchParams();
            if (query) params.append('q', query);
            if (country) params.append('country', country);
            
            try {
                const response = await fetch('/api/users/search?' + params.toString());
                const data = await response.json();
                const resultsDiv = document.getElementById('searchResults');
                if (data.results && data.results.length > 0) {
                    let html = '<div class="example"><strong>Found ' + data.count + ' users:</strong><br><br>';
                    data.results.forEach(user => {
                        html += user.id + '. ' + user.name + ' (' + user.email + ') - ' + user.country + ', Age: ' + user.age + ', Active: ' + user.active + '<br>';
                    });
                    html += '</div>';
                    resultsDiv.innerHTML = html;
                } else {
                    resultsDiv.innerHTML = '<p style="color: #999999;">No users found</p>';
                }
            } catch (error) {
                document.getElementById('searchResults').innerHTML = '<p style="color: #ff6666;">Error: ' + error.message + '</p>';
            }
        }
        
        function showTab(tabName) {
            document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
            
            event.target.classList.add('active');
            document.getElementById(tabName).classList.add('active');
        }
        
        async function loadMetrics() {
            try {
                const response = await fetch('/api/metrics');
                const data = await response.json();
                const metricsDiv = document.getElementById('metricsDisplay');
                
                if (data.metrics && data.metrics.length > 0) {
                    let html = '<div class="metrics-grid">';
                    data.metrics.forEach(metric => {
                        html += '<div class="metric-card">';
                        html += '<h4>' + metric.path + '</h4>';
                        html += '<div class="metric-value">' + metric.count + '</div>';
                        html += '<div class="metric-label">Requests</div>';
                        html += '<div style="margin-top: 15px;">';
                        html += '<div style="color: #cccccc; font-size: 0.9em;">Avg: ' + metric.avg_time_ms.toFixed(2) + 'ms</div>';
                        html += '<div style="color: #999999; font-size: 0.85em;">Min: ' + metric.min_time_ms.toFixed(2) + 'ms | Max: ' + metric.max_time_ms.toFixed(2) + 'ms</div>';
                        html += '</div>';
                        html += '</div>';
                    });
                    html += '</div>';
                    metricsDiv.innerHTML = html;
                } else {
                    metricsDiv.innerHTML = '<div class="example">No metrics available yet. Make some API requests first!</div>';
                }
            } catch (error) {
                document.getElementById('metricsDisplay').innerHTML = '<p style="color: #ff6666;">Error: ' + error.message + '</p>';
            }
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
	
	page := 1
	perPage := 10
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}
	
	allUsers := make([]User, 0, len(store.users))
	for _, user := range store.users {
		allUsers = append(allUsers, user)
	}
	
	// Sort users
	if sortBy != "" {
		switch sortBy {
		case "name":
			if order == "desc" {
				for i := 0; i < len(allUsers); i++ {
					for j := i + 1; j < len(allUsers); j++ {
						if allUsers[i].Name < allUsers[j].Name {
							allUsers[i], allUsers[j] = allUsers[j], allUsers[i]
						}
					}
				}
			} else {
				for i := 0; i < len(allUsers); i++ {
					for j := i + 1; j < len(allUsers); j++ {
						if allUsers[i].Name > allUsers[j].Name {
							allUsers[i], allUsers[j] = allUsers[j], allUsers[i]
						}
					}
				}
			}
		case "age":
			if order == "desc" {
				for i := 0; i < len(allUsers); i++ {
					for j := i + 1; j < len(allUsers); j++ {
						if allUsers[i].Age < allUsers[j].Age {
							allUsers[i], allUsers[j] = allUsers[j], allUsers[i]
						}
					}
				}
			} else {
				for i := 0; i < len(allUsers); i++ {
					for j := i + 1; j < len(allUsers); j++ {
						if allUsers[i].Age > allUsers[j].Age {
							allUsers[i], allUsers[j] = allUsers[j], allUsers[i]
						}
					}
				}
			}
		case "created":
			if order == "desc" {
				for i := 0; i < len(allUsers); i++ {
					for j := i + 1; j < len(allUsers); j++ {
						if allUsers[i].CreatedAt.Before(allUsers[j].CreatedAt) {
							allUsers[i], allUsers[j] = allUsers[j], allUsers[i]
						}
					}
				}
			} else {
				for i := 0; i < len(allUsers); i++ {
					for j := i + 1; j < len(allUsers); j++ {
						if allUsers[i].CreatedAt.After(allUsers[j].CreatedAt) {
							allUsers[i], allUsers[j] = allUsers[j], allUsers[i]
						}
					}
				}
			}
		}
	}
	
	total := len(allUsers)
	totalPages := (total + perPage - 1) / perPage
	
	start := (page - 1) * perPage
	end := start + perPage
	
	if start >= total {
		respondJSON(w, http.StatusOK, PaginatedResponse{
			Data:       []User{},
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		})
		return
	}
	
	if end > total {
		end = total
	}
	
	respondJSON(w, http.StatusOK, PaginatedResponse{
		Data:       allUsers[start:end],
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
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
		Name    string `json:"name"`
		Email   string `json:"email"`
		Age     int    `json:"age,omitempty"`
		Country string `json:"country,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	
	if input.Name == "" || input.Email == "" {
		respondError(w, http.StatusBadRequest, "Name and email are required")
		return
	}
	
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(strings.ToLower(input.Email)) {
		respondError(w, http.StatusBadRequest, "Invalid email format")
		return
	}
	
	if input.Age < 0 || input.Age > 150 {
		respondError(w, http.StatusBadRequest, "Age must be between 0 and 150")
		return
	}
	
	now := time.Now()
	store.mu.Lock()
	user := User{
		ID:        store.nextID,
		Name:      input.Name,
		Email:     input.Email,
		Age:       input.Age,
		Country:   input.Country,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
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
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	var input struct {
		Name    string `json:"name,omitempty"`
		Email   string `json:"email,omitempty"`
		Age     *int   `json:"age,omitempty"`
		Country string `json:"country,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	
	store.mu.Lock()
	user, exists := store.users[id]
	if !exists {
		store.mu.Unlock()
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
		if !emailRegex.MatchString(strings.ToLower(input.Email)) {
			store.mu.Unlock()
			respondError(w, http.StatusBadRequest, "Invalid email format")
			return
		}
		user.Email = input.Email
	}
	if input.Age != nil {
		if *input.Age < 0 || *input.Age > 150 {
			store.mu.Unlock()
			respondError(w, http.StatusBadRequest, "Age must be between 0 and 150")
			return
		}
		user.Age = *input.Age
	}
	if input.Country != "" {
		user.Country = input.Country
	}
	
	user.UpdatedAt = time.Now()
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
	stats.TotalUsers = len(store.users)
	stats.ActiveUsers = 0
	stats.UsersByCountry = make(map[string]int)
	
	for _, user := range store.users {
		if user.Active {
			stats.ActiveUsers++
		}
		if user.Country != "" {
			stats.UsersByCountry[user.Country]++
		}
	}
	
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

func searchUsers(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	
	query := strings.ToLower(r.URL.Query().Get("q"))
	country := r.URL.Query().Get("country")
	activeStr := r.URL.Query().Get("active")
	
	var results []User
	for _, user := range store.users {
		if query != "" {
			if !strings.Contains(strings.ToLower(user.Name), query) &&
			   !strings.Contains(strings.ToLower(user.Email), query) {
				continue
			}
		}
		
		if country != "" && user.Country != country {
			continue
		}
		
		if activeStr != "" {
			active, _ := strconv.ParseBool(activeStr)
			if user.Active != active {
				continue
			}
		}
		
		results = append(results, user)
	}
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
}

func batchCreateUsers(w http.ResponseWriter, r *http.Request) {
	var req BatchCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	
	if len(req.Users) == 0 {
		respondError(w, http.StatusBadRequest, "No users provided")
		return
	}
	
	if len(req.Users) > 100 {
		respondError(w, http.StatusBadRequest, "Maximum 100 users per batch")
		return
	}
	
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	
	var createdUsers []User
	store.mu.Lock()
	defer store.mu.Unlock()
	
	for _, userReq := range req.Users {
		if userReq.Name == "" || userReq.Email == "" {
			continue
		}
		
		if !emailRegex.MatchString(strings.ToLower(userReq.Email)) {
			continue
		}
		
		now := time.Now()
		user := User{
			ID:        store.nextID,
			Name:      userReq.Name,
			Email:     userReq.Email,
			Age:       userReq.Age,
			Country:   userReq.Country,
			Active:    true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		
		store.users[store.nextID] = user
		createdUsers = append(createdUsers, user)
		store.nextID++
	}
	
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"created": createdUsers,
		"count":   len(createdUsers),
	})
}

func batchDeleteUsers(w http.ResponseWriter, r *http.Request) {
	var req BatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	
	if len(req.IDs) == 0 {
		respondError(w, http.StatusBadRequest, "No IDs provided")
		return
	}
	
	store.mu.Lock()
	defer store.mu.Unlock()
	
	var deleted []int
	for _, id := range req.IDs {
		if _, exists := store.users[id]; exists {
			delete(store.users, id)
			deleted = append(deleted, id)
		}
	}
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": deleted,
		"count":   len(deleted),
	})
}

func activateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	store.mu.Lock()
	defer store.mu.Unlock()
	
	user, exists := store.users[id]
	if !exists {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	
	user.Active = true
	user.UpdatedAt = time.Now()
	store.users[id] = user
	
	respondJSON(w, http.StatusOK, user)
}

func deactivateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	store.mu.Lock()
	defer store.mu.Unlock()
	
	user, exists := store.users[id]
	if !exists {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	
	user.Active = false
	user.UpdatedAt = time.Now()
	store.users[id] = user
	
	respondJSON(w, http.StatusOK, user)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	totalUsers := len(store.users)
	activeUsers := 0
	for _, user := range store.users {
		if user.Active {
			activeUsers++
		}
	}
	store.mu.RUnlock()
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "healthy",
		"timestamp":    time.Now(),
		"uptime":       time.Since(store.stats.StartTime).String(),
		"total_users":  totalUsers,
		"active_users": activeUsers,
	})
}

func exportUsers(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	
	allUsers := make([]User, 0, len(store.users))
	for _, user := range store.users {
		allUsers = append(allUsers, user)
	}
	
	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=users.csv")
		
		fmt.Fprintf(w, "ID,Name,Email,Age,Country,Active,Created At,Updated At\n")
		for _, user := range allUsers {
			fmt.Fprintf(w, "%d,%s,%s,%d,%s,%t,%s,%s\n",
				user.ID, user.Name, user.Email, user.Age, user.Country, user.Active,
				user.CreatedAt.Format(time.RFC3339), user.UpdatedAt.Format(time.RFC3339))
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=users.json")
		json.NewEncoder(w).Encode(allUsers)
	}
}

func getUserAnalytics(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	
	totalUsers := len(store.users)
	activeUsers := 0
	inactiveUsers := 0
	byCountry := make(map[string]int)
	avgAge := 0
	ageSum := 0
	ageCount := 0
	
	for _, user := range store.users {
		if user.Active {
			activeUsers++
		} else {
			inactiveUsers++
		}
		
		if user.Country != "" {
			byCountry[user.Country]++
		}
		
		if user.Age > 0 {
			ageSum += user.Age
			ageCount++
		}
	}
	
	if ageCount > 0 {
		avgAge = ageSum / ageCount
	}
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_users":      totalUsers,
		"active_users":     activeUsers,
		"inactive_users":   inactiveUsers,
		"users_by_country": byCountry,
		"average_age":      avgAge,
		"timestamp":        time.Now(),
	})
}

func getMetrics(w http.ResponseWriter, r *http.Request) {
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()
	
	result := make([]map[string]interface{}, 0, len(metrics))
	for path, m := range metrics {
		result = append(result, map[string]interface{}{
			"path":         path,
			"count":        m.Count,
			"avg_time_ms":  float64(m.AverageTime.Microseconds()) / 1000.0,
			"min_time_ms":  float64(m.MinTime.Microseconds()) / 1000.0,
			"max_time_ms":  float64(m.MaxTime.Microseconds()) / 1000.0,
			"total_time_s": m.TotalTime.Seconds(),
		})
	}
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"metrics":   result,
		"timestamp": time.Now(),
	})
}

func trackPerformance(path string, duration time.Duration) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	
	m, exists := metrics[path]
	if !exists {
		m = &PerformanceMetrics{
			Path:    path,
			MinTime: duration,
			MaxTime: duration,
		}
		metrics[path] = m
	}
	
	m.Count++
	m.TotalTime += duration
	m.AverageTime = m.TotalTime / time.Duration(m.Count)
	
	if duration < m.MinTime {
		m.MinTime = duration
	}
	if duration > m.MaxTime {
		m.MaxTime = duration
	}
	
	store.mu.Lock()
	store.stats.RequestsByPath[path]++
	store.mu.Unlock()
}

func initTestData() {
	now := time.Now()
	store.users[1] = User{
		ID:        1,
		Name:      "Иван Петров",
		Email:     "ivan@example.com",
		Age:       30,
		Country:   "Russia",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.users[2] = User{
		ID:        2,
		Name:      "Мария Сидорова",
		Email:     "maria@example.com",
		Age:       25,
		Country:   "Russia",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.users[3] = User{
		ID:        3,
		Name:      "Петр Иванов",
		Email:     "petr@example.com",
		Age:       35,
		Country:   "Ukraine",
		Active:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.users[4] = User{
		ID:        4,
		Name:      "John Smith",
		Email:     "john@example.com",
		Age:       28,
		Country:   "USA",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.users[5] = User{
		ID:        5,
		Name:      "Anna Schmidt",
		Email:     "anna@example.com",
		Age:       32,
		Country:   "Germany",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.nextID = 6
	
	store.stats.UsersByCountry = make(map[string]int)
}
