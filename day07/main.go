package main

import "fmt"

func main() {
	alertsMap := map[string]int{"HighCPU": 3, "DiskFull": 1, "PodCrashLoop": 7}
	fmt.Println(alertsMap)
	for alertType, alertCount := range alertsMap {
		fmt.Printf("The alert %s occurs %v times \n", alertType, alertCount)
	}
	diskFull, ok := alertsMap["DiskFull"]
	if ok {
		fmt.Printf("DiskFull exists %v times \n", diskFull)
	}
	memoryLeak, exists := alertsMap["MemoryLeak"]
	if !exists {
		fmt.Printf("MemoryLeak does not exist its value is %v \n", memoryLeak)
	}
	alertsMap["HighCPU"]++
	delete(alertsMap, "DiskFull")
	fmt.Println("Printing the final Map")
	for alertType, alertCount := range alertsMap {
		fmt.Printf("The alert %s occurs %v times \n", alertType, alertCount)
	}
}
