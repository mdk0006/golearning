package main

import (
	"fmt"
	"log/slog"
	"os"
)

func checkServer(logger *slog.Logger, host, region string) error {
	logger.Info("starting health check ", "host", host, "region", region)
	if host == "db-01" {
		err := fmt.Errorf("connection refused")
		logger.Error("check failed", "host", host, "err", err)
		return err
	}
	logger.Info("check passed", "host", host)
	return nil
}
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger = logger.With("service", "health-checker", "version", "1.0")
	checkServer(logger, "web-01", "us-east1")
	checkServer(logger, "web-02", "eu-west1")
	checkServer(logger, "db-01", "us-east-1")
	checkServer(logger, "web-03", "us-east-4")
}
