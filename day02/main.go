package main

import "fmt"

func checkHealth(host string, threshold, cpuUsage float64) (string, bool) {
	if cpuUsage > threshold {
		return fmt.Sprintf("%s is unhealthy, cpu: %.1f", host, cpuUsage), false
	}
	return fmt.Sprintf("%s is healthy, cpu: %.1f", host, cpuUsage), true
}
func main() {
	threshold := 89.0
	cpuUsage := 90.1
	host := "prometheus"
	status, _ := checkHealth(host, threshold, cpuUsage)
	fmt.Println(status)
}
