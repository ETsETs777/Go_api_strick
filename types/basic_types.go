package types

import "fmt"

func DemoBasicTypes() {
	var intNum int = 42
	var int8Num int8 = 127
	var int16Num int16 = 32767
	var int32Num int32 = 2147483647
	var int64Num int64 = 9223372036854775807
	
	var uintNum uint = 42
	var uint8Num uint8 = 255
	var float32Num float32 = 3.14
	var float64Num float64 = 2.718281828
	var complexNum complex128 = 1 + 2i
	var str string = "Привет, Go!"
	var boolVal bool = true
	var byteVal byte = 'A'
	var runeVal rune = 'Я'
	
	fmt.Println("Целые числа:", intNum, int8Num, int16Num, int32Num, int64Num)
	fmt.Println("Беззнаковые:", uintNum, uint8Num)
	fmt.Println("Дробные:", float32Num, float64Num)
	fmt.Println("Комплексные:", complexNum)
	fmt.Println("Строка:", str)
	fmt.Println("Булево значение:", boolVal)
	fmt.Println("Байт:", byteVal, "Руна:", string(runeVal))
	
	shortVar := "Автоматическое определение типа"
	fmt.Println(shortVar)
	
	const Pi = 3.14159265359
	const Greeting = "Здравствуйте!"
	fmt.Printf("Константы: Pi = %v, Greeting = %s\n", Pi, Greeting)
}

func DemoStructs() {
	type Person struct {
		Name    string
		Age     int
		Email   string
		IsAdmin bool
	}
	
	person1 := Person{
		Name:    "Иван Петров",
		Age:     30,
		Email:   "ivan@example.com",
		IsAdmin: false,
	}
	
	person2 := Person{"Мария Сидорова", 25, "maria@example.com", true}
	
	config := struct {
		Host string
		Port int
	}{
		Host: "localhost",
		Port: 8080,
	}
	
	fmt.Printf("Person 1: %+v\n", person1)
	fmt.Printf("Person 2: %s, возраст %d\n", person2.Name, person2.Age)
	fmt.Printf("Config: %s:%d\n", config.Host, config.Port)
	
	type Address struct {
		City    string
		Country string
	}
	
	type Employee struct {
		Person  Person
		Address Address
		Salary  float64
	}
	
	emp := Employee{
		Person:  person1,
		Address: Address{"Москва", "Россия"},
		Salary:  75000.0,
	}
	
	fmt.Printf("Employee: %s работает в %s, зарплата: %.2f\n", 
		emp.Person.Name, emp.Address.City, emp.Salary)
}

func DemoArraysSlices() {
	var arr [5]int = [5]int{1, 2, 3, 4, 5}
	fmt.Println("Массив:", arr)
	
	arr2 := [...]string{"Go", "Python", "JavaScript"}
	fmt.Println("Массив строк:", arr2)
	
	slice := []int{10, 20, 30, 40, 50}
	fmt.Println("Срез:", slice)
	
	slice2 := make([]int, 5, 10)
	fmt.Printf("Срез с make: %v, len=%d, cap=%d\n", slice2, len(slice2), cap(slice2))
	
	slice = append(slice, 60, 70)
	fmt.Println("После append:", slice)
	
	subSlice := slice[1:4]
	fmt.Println("Под-срез [1:4]:", subSlice)
	
	slice3 := make([]int, len(slice))
	copy(slice3, slice)
	fmt.Println("Копия среза:", slice3)
}

func DemoMaps() {
	ages := make(map[string]int)
	ages["Alice"] = 25
	ages["Bob"] = 30
	ages["Charlie"] = 35
	
	fmt.Println("Map ages:", ages)
	
	capitals := map[string]string{
		"Россия":   "Москва",
		"США":      "Вашингтон",
		"Франция":  "Париж",
		"Германия": "Берлин",
	}
	
	fmt.Println("Столицы:", capitals)
	
	capital := capitals["Россия"]
	fmt.Println("Столица России:", capital)
	
	capital, exists := capitals["Италия"]
	if exists {
		fmt.Println("Столица Италии:", capital)
	} else {
		fmt.Println("Италия не найдена в карте")
	}
	
	delete(capitals, "США")
	fmt.Println("После удаления США:", capitals)
	
	fmt.Println("Все столицы:")
	for country, capital := range capitals {
		fmt.Printf("  %s: %s\n", country, capital)
	}
}
