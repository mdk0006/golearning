package main

import "fmt"

func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("caught panic: %v", r)
		}
	}()
	if b == 0 {
		panic("cannot divide by zero")
	}
	return a / b, nil
}
func connectDB(host string) {
	defer fmt.Printf("closing DB connection to %s \n", host)
	fmt.Printf("connecting to: %s\n", host)
	fmt.Println("running queries")
}
func main() {
	connectDB("postgres-01")
	result, err := safeDivide(10, 2)
	fmt.Println(result, err)
	result, err = safeDivide(10, 0)
	fmt.Println(result, err)
}
