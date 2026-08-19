package main

import (
	"fmt"
	"sync"
)

type HealthRegistry struct {
	status map[string]string
	mutex  sync.RWMutex
}

func (r *HealthRegistry) Set(host, status string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.status[host] = status
}

func (r *HealthRegistry) Get(host string) string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.status[host]

}

func main() {
	reg := &HealthRegistry{status: make(map[string]string)}
	var wg sync.WaitGroup
	servers := []string{"web01", "web-02", "db-01", "cache-01"}
	for _, server := range servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			if s == "db-01" {
				reg.Set(s, "CRITICAL")
			} else {
				reg.Set(s, "OK")
			}
		}(server)
	}
	wg.Wait()
	for _, server := range servers {
		fmt.Printf("%s: %s\n", server, reg.Get(server))
	}
}
