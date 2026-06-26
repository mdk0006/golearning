package main

import (
	"day11/healthcheck"
	"fmt"
)

func main() {
	s1 := healthcheck.Server{Name: "web-01", CPUUsage: 45.0}
	s2 := healthcheck.Server{Name: "web-02", CPUUsage: 95.5}
	fmt.Println(healthcheck.Check(s1))
	fmt.Println(healthcheck.Check(s2))
}
