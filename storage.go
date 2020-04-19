package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

// Storage stores last healthcheck results
type Storage struct {
	lock *sync.Mutex
}

// NewStorage creates new Storage object
func NewStorage() *Storage {
	return &Storage{
		lock: &sync.Mutex{},
	}
}

// StorageStatus is a state for storage item
type StorageStatus int

const (
	// StorageStateNoChange means healthcheck state didn't change
	StorageStateNoChange StorageStatus = iota
	// StorageStateAdded means healthcheck state has been jsut added
	StorageStateAdded
	// StorageStateUpdated means healthcheck state changes
	StorageStateUpdated
)

// StorageModelJSON is a root JSON model for Storage object
type StorageModelJSON struct {
	Urls map[string]*StorageItemModelJSON `json:"urls"`
}

// StorageItemModelJSON is a sub-item JSON model for StorageModelJSON
type StorageItemModelJSON struct {
	Time    time.Time `json:"time"`
	OK      bool      `json:"status"`
	Message string    `json:"message"`
}

// Update updates last healthcheck state
func (s *Storage) Update(r *HealthcheckResult) StorageStatus {
	s.lock.Lock()
	defer s.lock.Unlock()

	model := s.readModel()

	var status StorageStatus
	item, exists := model.Urls[r.Healthcheck.ID]
	if exists {
		if r.OK != item.OK {
			status = StorageStateUpdated
			if r.OK {
				log.Printf("[%s] host is now up", r.Healthcheck.Name)
			} else {
				log.Printf("[%s] host is now down: %s", r.Healthcheck.Name, r.Message)
			}
		} else {
			status = StorageStateNoChange
		}
	} else {
		status = StorageStateAdded
		item = &StorageItemModelJSON{}
		model.Urls[r.Healthcheck.ID] = item
	}

	item.OK = r.OK
	item.Time = r.Time
	item.Message = r.Message

	s.writeModel(model)

	return status
}

func (s *Storage) readModel() *StorageModelJSON {
	b, err := ioutil.ReadFile(StoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &StorageModelJSON{
				Urls: make(map[string]*StorageItemModelJSON),
			}
		}

		log.Printf("unable to load storage file \"%s\": %v", StoragePath, err)
		return &StorageModelJSON{
			Urls: make(map[string]*StorageItemModelJSON),
		}
	}

	var model StorageModelJSON
	err = json.Unmarshal(b, &model)
	if err != nil {
		log.Printf("unable to deserialize storage file \"%s\": %v", StoragePath, err)
		return &StorageModelJSON{
			Urls: make(map[string]*StorageItemModelJSON),
		}
	}

	return &model
}

func (s *Storage) writeModel(model *StorageModelJSON) {
	b, err := json.MarshalIndent(model, "", "    ")
	if err != nil {
		log.Printf("unable to serialize storage file \"%s\": %v", StoragePath, err)
		return
	}

	err = ioutil.WriteFile(StoragePath, b, 0666)
	if err != nil {
		log.Printf("unable to write storage file \"%s\": %v", StoragePath, err)
		return
	}
}
