package main

import "fmt"

func makeAlertCounter(service string) func() string {
	calls := 0
	return func() string {
		calls++
		return fmt.Sprintf("%s: alert #%d", service, calls)
	}
}
func makeRateLimiter(limit int) func() bool {
	count := 0
	return func() bool {
		count++
		return count <= limit
	}
}

func maxCPU(usages ...float64) float64 {
	max := usages[0]
	for _, usage := range usages {
		if usage > max {
			max = usage
		}
	}
	return max
}
func main() {
	webAlert := makeAlertCounter("web-01")
	dbAlert := makeAlertCounter("db-01")
	fmt.Println(webAlert())
	fmt.Println(webAlert())
	fmt.Println(dbAlert())
	dbLimiter := makeRateLimiter(5)
	for i := 0; i < 7; i++ {
		fmt.Println("dbLimter allowed", dbLimiter())
	}
	fmt.Printf("%.1f\n", maxCPU(45.2, 88.1, 23.5, 91.0, 67.3)) // 91.0

}
