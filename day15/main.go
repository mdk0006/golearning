package main

import (
	"encoding/json"
	"fmt"
)

type Server struct {
	Name     string  `json:"name"`
	CPUUsage float64 `json:"cpu_usage"`
	Region   string  `json:"region"`
	Healthy  bool    `json:"healthy"`
}

func main() {
	servers := []Server{
		{Name: "web-01", CPUUsage: 90, Region: "US-EAST1", Healthy: true},
		{Name: "web-02", CPUUsage: 20, Region: "US-EAST3", Healthy: true},
		{Name: "web-03", CPUUsage: 30, Region: "US-EAST2", Healthy: false},
	}
	input := `{"name":"db-01","region":"eu-west-1","cpu_usage":88.5,"healthy":false}`
	for _, server := range servers {
		data, err := json.MarshalIndent(server, "", "  ")
		if err != nil {
			fmt.Println("marshal failed:", err)
			continue
		}
		fmt.Println(string(data)) //because json marshal gives byte
	}
	var DBserver Server
	err := json.Unmarshal([]byte(input), &DBserver)
	if err != nil {
		fmt.Println("unmarshal failed:", err)
		return
	}
	fmt.Println("Name:", DBserver.Name)
	fmt.Println("CPUUsage:", DBserver.CPUUsage)
	fmt.Println("Region:", DBserver.Region)
	fmt.Println("Healthy:", DBserver.Healthy)
}
