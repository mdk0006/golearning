package main

import "fmt"

type Server struct {
	Hostname string
	IP       string
	Region   string
	Healthy  bool
}

func (s Server) Status() string {
	serverHealth := "Unhealthy"
	if s.Healthy {
		serverHealth = "Healthy"
	}
	return fmt.Sprintf("The %s [%s] - %s", s.Hostname, s.Region, serverHealth)
}

func FilterUnhealthy(servers []Server) []Server {
	serversHealth := []Server{}
	for _, server := range servers {
		if !server.Healthy {
			serversHealth = append(serversHealth, server)
		}
	}
	return serversHealth
}

func main() {
	servers := []Server{
		{Hostname: "web-01", IP: "10.0.0.1", Region: "us-east-1", Healthy: true},
		{Hostname: "web-02", IP: "10.0.0.2", Region: "us-east-1", Healthy: false},
		{Hostname: "web-03", IP: "10.0.0.3", Region: "us-east-2", Healthy: false},
	}
	for _, server := range servers {
		fmt.Println(server.Status())
	}
	fmt.Println("Printing only Unhealthy now")
	unhealthyServers := FilterUnhealthy(servers)
	for _, server := range unhealthyServers {
		fmt.Println(server.Status())
	}
}
