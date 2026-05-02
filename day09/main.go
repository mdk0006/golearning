package main

import "fmt"

type HealthChecker interface {
	CheckHealth() string
}

type WebServer struct {
	Name string
	URL  string
}
type Database struct {
	Name string
	Port int
}
type KubernetesNode struct {
	Name    string
	Healthy bool
}

func (w WebServer) CheckHealth() string {
	return fmt.Sprintf("Webserver %v: checking endpoint %v", w.Name, w.URL)
}
func (d Database) CheckHealth() string {
	return fmt.Sprintf("Database %v: checking port %v", d.Name, d.Port)
}
func (k KubernetesNode) CheckHealth() string {
	if k.Healthy {
		return fmt.Sprintf("KubernetesNode %v: node is ready", k.Name)
	} else {
		return fmt.Sprintf("KubernetesNode %v: node is not ready", k.Name)
	}
}

func runChecks(targets []HealthChecker) {
	for _, t := range targets {
		fmt.Println(t.CheckHealth())
	}
}
func main() {
	targets := []HealthChecker{
		WebServer{Name: "web-01", URL: "http://10.0.0.1"},
		Database{Name: "postgres-01", Port: 5432},
		KubernetesNode{Name: "node-01", Healthy: true},
		KubernetesNode{Name: "node-02", Healthy: false},
	}
	runChecks(targets)
}
