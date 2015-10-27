package main

import (
	"fmt"
	"log"
	"os"

	"ScoringDaemon/scoring"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: main conf")
		return
	}
	logFile, err := os.Create("/var/log/scoredaemon.log")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.SetOutput(logFile)
	scoring.Start(os.Args[1])
}
