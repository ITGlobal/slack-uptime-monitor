package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

// Healthchecker is a runtime healthchecker service
type Healthchecker struct {
	storage           *Storage
	notifier          *Notifier
	lastCheckTime     time.Time
	lastCheckTimeLock *sync.Mutex
}

// NewHealthchecker creates new Healthchecker object
func NewHealthchecker(storage *Storage, notifier *Notifier) *Healthchecker {
	s := &Healthchecker{
		storage:           storage,
		notifier:          notifier,
		lastCheckTimeLock: &sync.Mutex{},
	}

	return s
}

// Start starts healthchecker
func (s *Healthchecker) Start(group *sync.WaitGroup, stop chan interface{}) error {
	config, err := s.runOnce(true, nil)
	if err != nil {
		return err
	}

	s.startLoop(config, group, stop)
	s.startAPI(group, stop)

	return nil
}

func (s *Healthchecker) startLoop(config *Config, group *sync.WaitGroup, stop chan interface{}) {
	group.Add(1)
	t := time.NewTimer(LoopPeriod)

	log.Printf("will run healthchecks every %s", LoopPeriod)

	go func() {
		for range t.C {
			s.runOnce(false, config)
		}
	}()

	go func() {
		for range stop {
		}

		t.Stop()
		group.Done()
	}()
}

func (s *Healthchecker) startAPI(group *sync.WaitGroup, stop chan interface{}) {
	group.Add(1)
	server := &http.Server{Addr: ListenAddress}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		healthy := s.isHealthy()
		if healthy {
			w.WriteHeader(200)
			_, err := w.Write([]byte("OK"))
			if err != nil {
				panic(err)
			}
			return
		}

		w.WriteHeader(500)
		_, err := w.Write([]byte("Unhealthy"))
		if err != nil {
			panic(err)
		}
	})

	go func() {
		log.Printf("listening on \"%s\"\n", server.Addr)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP: could not listen on \"%s\": %v\n", server.Addr, err)
		}

		group.Done()
	}()

	go func() {
		for range stop {
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		err := server.Shutdown(ctx)
		if err != nil {
			log.Fatalf("HTTP: could not gracefully shutdown the server: %v\n", err)
		}
	}()
}

func (s *Healthchecker) runOnce(initial bool, prevConfig *Config) (*Config, error) {
	config, err := LoadConfig()
	if err != nil {
		if prevConfig == nil || !initial {
			return prevConfig, err
		}

		log.Printf("unable to load config file: %v", err)
		log.Printf("will use last valid config instead")
		config = prevConfig
	}

	g := &sync.WaitGroup{}
	for _, h := range config.Healthchecks {
		g.Add(1)

		go func() {
			r := ExecuteHealthcheck(h)
			st := s.storage.Update(r)

			shouldNotify := initial && st == StorageStateUpdated || st != StorageStateNoChange
			if shouldNotify {
				s.notifier.Notify(r)
			}

			g.Done()
		}()
	}
	g.Wait()

	s.lastCheckTimeLock.Lock()
	defer s.lastCheckTimeLock.Unlock()
	s.lastCheckTime = time.Now().UTC()

	return config, nil
}

func (s *Healthchecker) isHealthy() bool {
	s.lastCheckTimeLock.Lock()
	defer s.lastCheckTimeLock.Unlock()

	now := time.Now().UTC()
	ago := now.Sub(s.lastCheckTime)

	if ago < 2*LoopPeriod {
		return true
	}

	log.Printf("HTTP: unhealthy, no checks since $s", s.lastCheckTime)
	return false
}
