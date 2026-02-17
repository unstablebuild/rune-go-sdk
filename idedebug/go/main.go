package main

import "fmt"

// Add adds two integers.
func Add(a, b int) int {
	return a + b
}

func main() {
	x := 10
	y := 20
	sum := Add(x, y)
	fmt.Println("Sum:", sum)
}
