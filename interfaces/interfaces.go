package interfaces

import (
	"errors"
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
	Name() string
}

type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

func (r Rectangle) Name() string {
	return "Прямоугольник"
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func (c Circle) Name() string {
	return "Круг"
}

type Triangle struct {
	A, B, C float64
}

func (t Triangle) Area() float64 {
	s := (t.A + t.B + t.C) / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

func (t Triangle) Perimeter() float64 {
	return t.A + t.B + t.C
}

func (t Triangle) Name() string {
	return "Треугольник"
}

func DemoInterfaces() {
	shapes := []Shape{
		Rectangle{Width: 10, Height: 5},
		Circle{Radius: 7},
		Triangle{A: 3, B: 4, C: 5},
	}
	
	fmt.Println("Геометрические фигуры:")
	for _, shape := range shapes {
		printShapeInfo(shape)
	}
	
	var anything interface{}
	anything = 42
	fmt.Printf("Пустой интерфейс (int): %v\n", anything)
	anything = "строка"
	fmt.Printf("Пустой интерфейс (string): %v\n", anything)
	
	str, ok := anything.(string)
	if ok {
		fmt.Printf("Type assertion успешен: %s\n", str)
	}
	
	describeType(42)
	describeType("Hello")
	describeType(3.14)
	describeType(true)
	describeType(Rectangle{Width: 5, Height: 3})
}

func printShapeInfo(s Shape) {
	fmt.Printf("  %s: Площадь=%.2f, Периметр=%.2f\n", 
		s.Name(), s.Area(), s.Perimeter())
}

func describeType(i interface{}) {
	switch v := i.(type) {
	case int:
		fmt.Printf("Это целое число: %d\n", v)
	case string:
		fmt.Printf("Это строка: %s\n", v)
	case float64:
		fmt.Printf("Это число с плавающей точкой: %.2f\n", v)
	case bool:
		fmt.Printf("Это булево значение: %t\n", v)
	case Shape:
		fmt.Printf("Это фигура: %s\n", v.Name())
	default:
		fmt.Printf("Неизвестный тип: %T\n", v)
	}
}

func DemoErrorHandling() {
	result, err := divide(10, 2)
	if err != nil {
		fmt.Println("Ошибка:", err)
	} else {
		fmt.Printf("10 / 2 = %.2f\n", result)
	}
	
	result, err = divide(10, 0)
	if err != nil {
		fmt.Println("Ошибка:", err)
	} else {
		fmt.Printf("10 / 0 = %.2f\n", result)
	}
	
	err = validateAge(-5)
	if err != nil {
		fmt.Println("Валидация:", err)
		
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			fmt.Printf("  Поле: %s, Значение: %v\n", 
				validationErr.Field, validationErr.Value)
		}
	}
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("деление на ноль")
	}
	return a / b, nil
}

type ValidationError struct {
	Field string
	Value interface{}
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("ошибка валидации поля '%s': %s (значение: %v)", 
		e.Field, e.Msg, e.Value)
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{
			Field: "age",
			Value: age,
			Msg:   "возраст не может быть отрицательным",
		}
	}
	if age > 150 {
		return &ValidationError{
			Field: "age",
			Value: age,
			Msg:   "возраст слишком большой",
		}
	}
	return nil
}
