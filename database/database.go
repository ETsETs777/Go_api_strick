package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func DemoDatabase() {
	db, err := sql.Open("sqlite3", "./demo.db")
	if err != nil {
		log.Printf("Ошибка открытия БД: %v\n", err)
		return
	}
	defer db.Close()
	defer os.Remove("./demo.db")
	
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		age INTEGER
	);`
	
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Printf("Ошибка создания таблицы: %v\n", err)
		return
	}
	fmt.Println("Таблица users создана")
	
	users := []User{
		{Name: "Иван Петров", Email: "ivan@example.com", Age: 30},
		{Name: "Мария Сидорова", Email: "maria@example.com", Age: 25},
		{Name: "Петр Иванов", Email: "petr@example.com", Age: 35},
	}
	
	for _, user := range users {
		insertUser(db, user)
	}
	
	fmt.Println("\nВсе пользователи:")
	allUsers := getAllUsers(db)
	for _, user := range allUsers {
		fmt.Printf("  ID: %d, Имя: %s, Email: %s, Возраст: %d\n", 
			user.ID, user.Name, user.Email, user.Age)
	}
	
	fmt.Println("\nПользователь с ID=2:")
	user, err := getUserByID(db, 2)
	if err != nil {
		log.Printf("Ошибка получения пользователя: %v\n", err)
	} else {
		fmt.Printf("  %+v\n", user)
	}
	
	fmt.Println("\nОбновление пользователя ID=1:")
	updateUser(db, 1, User{Name: "Иван Петров", Email: "ivan.new@example.com", Age: 31})
	
	fmt.Println("\nУдаление пользователя ID=3:")
	deleteUser(db, 3)
	
	fmt.Println("\nДемонстрация транзакции:")
	demoTransaction(db)
	
	fmt.Println("\nЭкспорт в JSON:")
	exportToJSON(db, "users.json")
}

func insertUser(db *sql.DB, user User) {
	insertSQL := `INSERT INTO users (name, email, age) VALUES (?, ?, ?)`
	result, err := db.Exec(insertSQL, user.Name, user.Email, user.Age)
	if err != nil {
		log.Printf("Ошибка вставки: %v\n", err)
		return
	}
	
	id, _ := result.LastInsertId()
	fmt.Printf("Вставлен пользователь с ID: %d\n", id)
}

func getAllUsers(db *sql.DB) []User {
	rows, err := db.Query("SELECT id, name, email, age FROM users")
	if err != nil {
		log.Printf("Ошибка запроса: %v\n", err)
		return nil
	}
	defer rows.Close()
	
	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age)
		if err != nil {
			log.Printf("Ошибка сканирования: %v\n", err)
			continue
		}
		users = append(users, user)
	}
	
	return users
}

func getUserByID(db *sql.DB, id int) (*User, error) {
	query := "SELECT id, name, email, age FROM users WHERE id = ?"
	row := db.QueryRow(query, id)
	
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Age)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func updateUser(db *sql.DB, id int, user User) {
	updateSQL := `UPDATE users SET name = ?, email = ?, age = ? WHERE id = ?`
	_, err := db.Exec(updateSQL, user.Name, user.Email, user.Age, id)
	if err != nil {
		log.Printf("Ошибка обновления: %v\n", err)
		return
	}
	fmt.Printf("Пользователь ID=%d обновлен\n", id)
}

func deleteUser(db *sql.DB, id int) {
	deleteSQL := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(deleteSQL, id)
	if err != nil {
		log.Printf("Ошибка удаления: %v\n", err)
		return
	}
	fmt.Printf("Пользователь ID=%d удален\n", id)
}

func demoTransaction(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v\n", err)
		return
	}
	
	_, err = tx.Exec("INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
		"Анна Кузнецова", "anna@example.com", 28)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка в транзакции: %v\n", err)
		return
	}
	
	_, err = tx.Exec("INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
		"Дмитрий Смирнов", "dmitry@example.com", 32)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка в транзакции: %v\n", err)
		return
	}
	
	err = tx.Commit()
	if err != nil {
		log.Printf("Ошибка коммита: %v\n", err)
		return
	}
	
	fmt.Println("Транзакция успешно выполнена (2 пользователя добавлены)")
}

func exportToJSON(db *sql.DB, filename string) {
	users := getAllUsers(db)
	
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		log.Printf("Ошибка маршалинга JSON: %v\n", err)
		return
	}
	
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Printf("Ошибка записи файла: %v\n", err)
		return
	}
	
	fmt.Printf("Данные экспортированы в %s\n", filename)
	defer os.Remove(filename)
}
