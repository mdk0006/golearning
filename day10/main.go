package main

import "fmt"

func main() {
	servers := []string{
		"web-01",
		"web-02",
		"web-03",
		"db-01",
		"db-02",
	}
	serversChannel := make(chan string, len(servers))
	//1.Launching all go routines
	for _, v := range servers {
		go func(name string) {
			serversChannel <- name + ": OK"
		}(v)
	}
	//2. Collecting all results
	for i := 0; i < len(servers); i++ {
		fmt.Println(<-serversChannel)
	}
	fmt.Println("All checks complete")
}
