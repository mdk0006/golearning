package main 
import "fmt"
var unhealthy bool
var errorCount int
var errorMsg string
var port int =8080
var host = "prometheus.internal"
func main() {
    healthy := true
	cpuUsage := 99.5
	fmt.Printf("The cpu usage is %v \n",cpuUsage)
	fmt.Printf("If cpu is healthy: %v \n",healthy)
	fmt.Printf("The port for prometheus is %v \n",port)
	fmt.Printf("The host name is %v \n",host)

fmt.Printf("unhealthy: %v, errorCount: %v, errorMsg: %q\n", unhealthy, errorCount, errorMsg)
}

