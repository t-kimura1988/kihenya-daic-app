package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	daicapp "github.com/t-kimura1988/kihenya-daic-app"
)

func main() {
	port := flag.Int("port", 50054, "The server port")

	if err := godotenv.Overload(); err != nil {
		logrus.WithError(err).Warn("Error loading .env file")
	}

	flag.Parse()
	daicapp.InitLogging()

	serverPort := *port
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			serverPort = p
		}
	}

	addr := fmt.Sprintf(":%d", serverPort)
	if err := daicapp.ListenAndServe(addr); err != nil {
		logrus.WithError(err).Fatal("failed to serve")
	}
}
