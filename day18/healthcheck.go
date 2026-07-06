package healthcheck

func CheckCPU(usage float64) string {
	if usage > 90 {
		return "CRITICAL"
	}
	return "OK"
}

func CheckDisk(freeGB float64) string {
	if freeGB < 10 {
		return "WARNING"
	}
	return "OK"
}
