package generics

import "fmt"

func DemoGenerics() {
	fmt.Println("Минимум из чисел:", Min(10, 5))
	fmt.Println("Минимум из строк:", Min("zebra", "apple"))
	
	fmt.Println("Максимум из чисел:", Max(3.14, 2.71))
	fmt.Println("Максимум из строк:", Max("Go", "Python"))
	
	intSlice := []int{5, 2, 8, 1, 9, 3}
	fmt.Println("Слайс int:", intSlice)
	fmt.Println("Минимум:", SliceMin(intSlice))
	fmt.Println("Максимум:", SliceMax(intSlice))
	
	stringSlice := []string{"banana", "apple", "cherry"}
	fmt.Println("Слайс string:", stringSlice)
	fmt.Println("Минимум:", SliceMin(stringSlice))
	fmt.Println("Максимум:", SliceMax(stringSlice))
	
	intStack := NewStack[int]()
	intStack.Push(1)
	intStack.Push(2)
	intStack.Push(3)
	fmt.Printf("Стек int, размер: %d\n", intStack.Size())
	fmt.Println("Pop:", intStack.Pop())
	fmt.Println("Pop:", intStack.Pop())
	
	stringStack := NewStack[string]()
	stringStack.Push("Hello")
	stringStack.Push("World")
	fmt.Printf("Стек string, размер: %d\n", stringStack.Size())
	fmt.Println("Pop:", stringStack.Pop())
	
	intMap := NewGenericMap[string, int]()
	intMap.Set("one", 1)
	intMap.Set("two", 2)
	intMap.Set("three", 3)
	
	fmt.Println("Значение 'two':", intMap.Get("two"))
	fmt.Println("Ключи:", intMap.Keys())
	
	fmt.Println("Сумма int:", Sum([]int{1, 2, 3, 4, 5}))
	fmt.Println("Сумма float64:", Sum([]float64{1.1, 2.2, 3.3}))
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func SliceMin[T Ordered](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func SliceMax[T Ordered](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: []T{}}
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() T {
	if len(s.items) == 0 {
		var zero T
		return zero
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}

func (s *Stack[T]) Size() int {
	return len(s.items)
}

type GenericMap[K comparable, V any] struct {
	data map[K]V
}

func NewGenericMap[K comparable, V any]() *GenericMap[K, V] {
	return &GenericMap[K, V]{data: make(map[K]V)}
}

func (m *GenericMap[K, V]) Set(key K, value V) {
	m.data[key] = value
}

func (m *GenericMap[K, V]) Get(key K) V {
	return m.data[key]
}

func (m *GenericMap[K, V]) Keys() []K {
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func Sum[T Number](numbers []T) T {
	var sum T
	for _, n := range numbers {
		sum += n
	}
	return sum
}
