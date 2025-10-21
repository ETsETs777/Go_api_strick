package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go-showcase/advanced"
	"go-showcase/concurrency"
	"go-showcase/generics"
	"go-showcase/interfaces"
	"go-showcase/reflection"
	"go-showcase/server"
	"go-showcase/types"
)

func main() {
	fmt.Println("=== GO Language Showcase ===")
	fmt.Println("Демонстрация всех возможностей Go\n")

	fmt.Println("--- 1. Типы данных ---")
	types.DemoBasicTypes()
	types.DemoStructs()
	types.DemoArraysSlices()
	types.DemoMaps()
	fmt.Println()

	fmt.Println("--- 2. Интерфейсы ---")
	interfaces.DemoInterfaces()
	interfaces.DemoErrorHandling()
	fmt.Println()

	fmt.Println("--- 3. Конкурентность (Goroutines & Channels) ---")
	concurrency.DemoGoroutines()
	concurrency.DemoChannels()
	concurrency.DemoSelect()
	concurrency.DemoWorkerPool()
	concurrency.DemoMutex()
	fmt.Println()

	fmt.Println("--- 4. Generics ---")
	generics.DemoGenerics()
	fmt.Println()

	fmt.Println("--- 5. Рефлексия ---")
	reflection.DemoReflection()
	fmt.Println()

	fmt.Println("--- 6. Defer, Panic, Recover ---")
	deferPanicRecover()
	fmt.Println()

	fmt.Println("--- 7. Работа с файлами ---")
	demoFileOperations()
	fmt.Println()

	fmt.Println("--- 8. Продвинутые паттерны конкурентности ---")
	advanced.DemoAdvancedPatterns()
	advanced.DemoCache()
	fmt.Println()

	fmt.Println("--- 9. Context ---")
	demoContext()
	fmt.Println()

	fmt.Println("--- 10. HTTP Server с Advanced Features ---")
	fmt.Println("Сервер включает:")
	fmt.Println("  ✅ REST API")
	fmt.Println("  ✅ WebSocket (ws://localhost:8080/ws)")
	fmt.Println("  ✅ Rate Limiting (10 req/s)")
	fmt.Println("  ✅ CORS")
	fmt.Println("  ✅ Security Headers")
	fmt.Println("  ✅ Graceful Shutdown")
	fmt.Println("  ✅ Structured Logging")
	fmt.Println("\nОткройте браузер: http://localhost:8080")
	fmt.Println("Нажмите Ctrl+C для graceful остановки сервера\n")
	
	server.StartServer()
}

func deferPanicRecover() {
	defer fmt.Println("Это выполнится последним (defer)")
	defer fmt.Println("3")
	defer fmt.Println("2")
	defer fmt.Println("1")
	
	fmt.Println("Начало функции")
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Восстановление от panic: %v\n", r)
			}
		}()
		
		fmt.Println("Собираюсь вызвать panic...")
		panic("Упс! Что-то пошло не так!")
	}()
	
	fmt.Println("Программа продолжает работать после recover")
}

func demoFileOperations() {
	filename := "test_file.txt"
	
	content := []byte("Привет, Go!\nЭто демонстрация работы с файлами.\n")
	err := os.WriteFile(filename, content, 0644)
	if err != nil {
		log.Printf("Ошибка записи файла: %v\n", err)
		return
	}
	fmt.Printf("Записан файл: %s\n", filename)
	
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Ошибка чтения файла: %v\n", err)
		return
	}
	fmt.Printf("Содержимое файла:\n%s\n", string(data))
	
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Ошибка открытия файла: %v\n", err)
		return
	}
	defer file.Close()
	
	if _, err := file.WriteString("Дополнительная строка!\n"); err != nil {
		log.Printf("Ошибка добавления в файл: %v\n", err)
		return
	}
	fmt.Println("Данные добавлены в файл")
	
	defer func() {
		if err := os.Remove(filename); err != nil {
			log.Printf("Ошибка удаления файла: %v\n", err)
		} else {
			fmt.Printf("Файл %s удален\n", filename)
		}
	}()
}

func demoContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	select {
	case <-time.After(1 * time.Second):
		fmt.Println("Операция завершена за 1 секунду")
	case <-ctx.Done():
		fmt.Println("Таймаут истек:", ctx.Err())
	}
	
	ctx2, cancel2 := context.WithCancel(context.Background())
	
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel2()
	}()
	
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("Операция завершена")
	case <-ctx2.Done():
		fmt.Println("Операция отменена:", ctx2.Err())
	}
}

