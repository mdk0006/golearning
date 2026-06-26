package healthcheck

import "fmt"

type Server struct {
	Name     string
	CPUUsage float64
}

func Check(s Server) string {
	if s.CPUUsage > 90 {
		return fmt.Sprintf("%s: CRITICAL - CPU %.1f", s.Name, s.CPUUsage)
	}
	return fmt.Sprintf("%s: OK", s.Name)
}
