package storage_driver

import (
	"errors"
	"sync"

	"github.com/gasparian/go-api-template/pkg/storage"
)

var (
	errUserNotFound = errors.New("user not found")
)

// KVDriver ...
type KVDriver struct {
	mx    sync.RWMutex
	users map[string]*storage.UserRecord
}

// NewKVDriver ...
func NewKVDriver() *KVDriver {
	return &KVDriver{
		users: make(map[string]*storage.UserRecord),
	}
}

// Ping ...
func (s *KVDriver) Ping() error {
	// should check that underlying systems work
	return nil
}

// Get ...
func (s *KVDriver) Get(userID string) (storage.UserRecord, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return storage.UserRecord{}, errUserNotFound
	}
	return *user, nil
}

// Set ...
func (s *KVDriver) Set(userID, lastItemID, visitorID string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	userRecord, ok := s.users[userID]
	if !ok {
		userRecord = &storage.UserRecord{
			UserID:         userID,
			TotalItemsSeen: 1,
			LastItemID:     lastItemID,
			VisitorID:      visitorID,
		}
	} else {
		userRecord.LastItemID = lastItemID
		userRecord.TotalItemsSeen += 1
	}
	s.users[userID] = userRecord
	return nil
}
