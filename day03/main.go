package main

import "fmt"

func printStatus(server, status string) {
	switch status {
	case "healthy":
		fmt.Printf("The %s is healthy\n", server)
	case "degraded":
		fmt.Printf("The %s is degraded\n", server)
	case "down":
		fmt.Printf("The %s is down\n", server)
	default:
		fmt.Printf("The %s has unknown status", server)
	}
}

func main() {
	statuses := []string{"healthy", "degraded", "down"}
	servers := []string{"node-01", "node-02", "node-03"}
	for i, server := range servers {
		printStatus(server, statuses[i])
	}
}
