package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func checkServer(logger *slog.Logger, host, region string) error {
	logger.Info("running health check ", "host", host, "region", region)
	if host == "db-01" {
		err := fmt.Errorf("health check failed")
		logger.Error("check failed", "host", host, "err", err)
		return err
	}
	logger.Info("check passed", "host", host)
	return nil
}

// go run main.go --host web-01 --region us-east-1

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	host := flag.String("host", "localhost", "hostname to check")
	region := flag.String("region", "us-east-1", "region of host")
	flag.Parse()
	err := checkServer(logger, *host, *region)
	if err != nil {
		os.Exit(1)
	}

	// fmt.Println(*host)
	// fmt.Println(*region)
}
