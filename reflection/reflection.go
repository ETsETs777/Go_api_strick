package reflection

import (
	"fmt"
	"reflect"
)

func DemoReflection() {
	var x float64 = 3.14
	fmt.Println("Тип:", reflect.TypeOf(x))
	fmt.Println("Значение:", reflect.ValueOf(x))
	
	type Person struct {
		Name  string `json:"name" validate:"required"`
		Age   int    `json:"age" validate:"min=0,max=150"`
		Email string `json:"email"`
	}
	
	p := Person{
		Name:  "Иван",
		Age:   30,
		Email: "ivan@example.com",
	}
	
	inspectStruct(p)
	
	fmt.Println("\nИзменение значений через рефлексию:")
	modifyValue(&x)
	fmt.Println("Новое значение x:", x)
	
	modifyStruct(&p)
	fmt.Printf("Измененная структура: %+v\n", p)
	
	fmt.Println("\nВызов методов через рефлексию:")
	calculator := Calculator{}
	callMethod(calculator, "Add", 5, 3)
	callMethod(calculator, "Multiply", 4, 7)
}

func inspectStruct(s interface{}) {
	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)
	
	fmt.Printf("\nИнспекция структуры: %s\n", t.Name())
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		
		fmt.Printf("  Поле: %s\n", field.Name)
		fmt.Printf("    Тип: %s\n", field.Type)
		fmt.Printf("    Значение: %v\n", value.Interface())
		
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			fmt.Printf("    JSON тег: %s\n", jsonTag)
		}
		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			fmt.Printf("    Validate тег: %s\n", validateTag)
		}
	}
}

func modifyValue(x interface{}) {
	v := reflect.ValueOf(x)
	
	if v.Kind() != reflect.Ptr {
		fmt.Println("Нужен указатель для изменения значения")
		return
	}
	
	v = v.Elem()
	
	if !v.CanSet() {
		fmt.Println("Значение нельзя изменить")
		return
	}
	
	if v.Kind() == reflect.Float64 {
		v.SetFloat(6.28)
	}
}

func modifyStruct(s interface{}) {
	v := reflect.ValueOf(s).Elem()
	
	ageField := v.FieldByName("Age")
	if ageField.IsValid() && ageField.CanSet() {
		ageField.SetInt(35)
	}
	
	emailField := v.FieldByName("Email")
	if emailField.IsValid() && emailField.CanSet() {
		emailField.SetString("new_email@example.com")
	}
}

type Calculator struct{}

func (c Calculator) Add(a, b int) int {
	result := a + b
	fmt.Printf("  Add(%d, %d) = %d\n", a, b, result)
	return result
}

func (c Calculator) Multiply(a, b int) int {
	result := a * b
	fmt.Printf("  Multiply(%d, %d) = %d\n", a, b, result)
	return result
}

func callMethod(obj interface{}, methodName string, args ...interface{}) {
	v := reflect.ValueOf(obj)
	method := v.MethodByName(methodName)
	
	if !method.IsValid() {
		fmt.Printf("Метод %s не найден\n", methodName)
		return
	}
	
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	
	method.Call(in)
}
