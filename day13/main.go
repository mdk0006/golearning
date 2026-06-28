package main

import (
	"fmt"
	"net/http"
	"time"
)

func checkURL(client *http.Client, url string) string {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Sprintf("ERROR: %s - %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return fmt.Sprintf("OK: %s", url)
	}
	return fmt.Sprintf("FAIL: %s - status %d", url, resp.StatusCode)
}

func main() {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	urls := []string{"https://httpbin.org/status/200", "https://httpbin.org/status/500", "https://httpbin.org/delay/5"}
	for _, url := range urls {
		fmt.Println(checkURL(client, url))
	}
}
