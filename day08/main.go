package main

import (
	"fmt"
)

type Server struct {
	Name     string
	Host     string
	CPUUsage float64
}

func checkHealth(s Server) error {
	if s.Host == "" {
		return fmt.Errorf("server %s: host is not configured", s.Name)

	}
	if s.CPUUsage > 90.0 {
		return fmt.Errorf("server %s: CPU usage %.1f exceeds threshold", s.Name, s.CPUUsage)

	}
	return nil
}

func main() {
	server1 := Server{Name: "web-01", Host: "10.0.0.1", CPUUsage: 45.0}
	server2 := Server{Name: "web-02", Host: "", CPUUsage: 30.0}
	server3 := Server{Name: "web-03", Host: "10.0.0.3", CPUUsage: 95.5}
	err1 := checkHealth(server1)
	err2 := checkHealth(server2)
	err3 := checkHealth(server3)
	if err1 != nil {
		fmt.Println("ALERT:", err1)
	} else {
		fmt.Printf("OK: %v is healthy \n", server1.Name)
	}
	if err2 != nil {
		fmt.Println("ALERT:", err2)
	} else {
		fmt.Printf("OK: %v is healthy \n", server2.Name)

	}
	if err3 != nil {
		fmt.Println("ALERT:", err3)
	} else {
		fmt.Printf("OK: %v is healthy \n", server3.Name)

	}

}
