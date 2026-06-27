package main

import (
	"bufio"
	"fmt"
	"os"
)

func serverHealth(name string) string {
	return name + ": OK"
}

func main() {
	file, err := os.Open("servers.txt")
	if err != nil {
		fmt.Println("error opening file:", err)
		return
	}
	defer file.Close()
	report, err := os.Create("report.txt")
	if err != nil {
		fmt.Println("error opening file:", err)
		return
	}
	defer report.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(report, "%s\n", serverHealth(line))
	}
	fmt.Println("Report written to report.txt")
}
