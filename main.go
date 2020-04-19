package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"
)

var (
	varPath string

	// ConfigPath stores path to config file
	ConfigPath string

	// StoragePath stores path to storage file
	StoragePath string

	// LoopPeriod stores delay between two healthcheck routines
	LoopPeriod time.Duration

	// SlackToken stores access token for Slack
	SlackToken string

	// SlackUsername stores custom username for Slack
	SlackUsername string

	// ListenAddress stores HTTP API endpoint
	ListenAddress string
)

func init() {
	flag.StringVar(&varPath, "d", "", "path to data directory (defaults to $VAR_DIR)")
	flag.StringVar(&ConfigPath, "c", "", "path to config file (defaults to $VAR_DIR/config.yaml)")
	flag.StringVar(&ListenAddress, "e", "0.0.0.0:5000", "endpoint for http api")
	flag.StringVar(&SlackToken, "t", "", "slack access token (defaults to $SLACK_TOKEN)")
	flag.StringVar(&SlackUsername, "u", "UptimeMonitor", "slack username (defaults to $SLACK_USERNAME)")
	flag.DurationVar(&LoopPeriod, "p", time.Minute, "period between two healthcheck routines")
}

func initialize() error {
	flag.Parse()

	// varPath
	if varPath == "" {
		varPath = os.Getenv("VAR_DIR")
		if varPath == "" {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			varPath = path.Join(wd, "var")
		}
	}

	err := os.MkdirAll(varPath, 0666)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// ConfigPath
	if ConfigPath == "" {
		ConfigPath = path.Join(varPath, "config.yaml")
	}

	// ListenAddress
	if ListenAddress == "" {
		ListenAddress = os.Getenv("LISTEN_ADDRESS")
		if ListenAddress == "" {
			ListenAddress = "0.0.0.0:5000"
		}
	}

	// SlackToken
	if SlackToken == "" {
		SlackToken = os.Getenv("SLACK_TOKEN")
		if SlackToken == "" {
			return fmt.Errorf("Slack access token is not set")
		}
	}

	// SlackUsername
	if SlackUsername == "" {
		SlackUsername = os.Getenv("SLACK_USERNAME")
		if SlackUsername == "" {
			SlackUsername = "UptimeMonitor"
		}
	}

	// LoopPeriod
	if LoopPeriod.Seconds() <= 1 {
		LoopPeriod = time.Minute
	}

	// StoragePath
	StoragePath = path.Join(varPath, "state.json")

	return nil
}

func main() {
	err := initialize()
	if err != nil {
		panic(err)
	}

	notifier, err := NewNotifier()
	if err != nil {
		panic(err)
	}
	storage := NewStorage()
	healthchecker := NewHealthchecker(storage, notifier)

	stop := make(chan interface{})
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		log.Printf("sigint received, shutting down")
		close(stop)
	}()

	group := &sync.WaitGroup{}

	err = healthchecker.Start(group, stop)
	if err != nil {
		panic(err)
	}

	log.Printf("up and running")
	group.Wait()
	log.Printf("good bye")
}
