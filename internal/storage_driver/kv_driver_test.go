package storage_driver

import (
	"fmt"
	"testing"

	"github.com/gasparian/go-api-template/pkg/storage"
)

// TestNewKVDriver verifies that a new KVDriver is initialized correctly.
func TestNewKVDriver(t *testing.T) {
	driver := NewKVDriver()
	if driver == nil {
		t.Fatal("Expected NewKVDriver to return a non-nil driver")
	}

	if driver.users == nil {
		t.Fatal("Expected users map to be initialized")
	}

	if len(driver.users) != 0 {
		t.Fatalf("Expected users map to be empty, got %d entries", len(driver.users))
	}
}

// TestPing verifies that the Ping method returns no error.
func TestPing(t *testing.T) {
	driver := NewKVDriver()
	err := driver.Ping()
	if err != nil {
		t.Errorf("Expected Ping to return nil, got %v", err)
	}
}

// TestSetAndGet verifies the Set and Get methods for adding and retrieving user records.
func TestSetAndGet(t *testing.T) {
	driver := NewKVDriver()

	userID := "user123"
	lastItemID := "item-1"
	visitorID := "visitor-456"

	// Initially, the user should not exist.
	_, err := driver.Get(userID)
	if err == nil {
		t.Fatalf("Expected error when getting non-existent user, got nil")
	}
	if err != errUserNotFound {
		t.Fatalf("Expected errUserNotFound, got %v", err)
	}

	// Set a new user.
	err = driver.Set(userID, lastItemID, visitorID)
	if err != nil {
		t.Fatalf("Set returned an unexpected error: %v", err)
	}

	// Retrieve the newly set user.
	user, err := driver.Get(userID)
	if err != nil {
		t.Fatalf("Get returned an unexpected error: %v", err)
	}

	expectedUser := storage.UserRecord{
		UserID:         userID,
		TotalItemsSeen: 1,
		LastItemID:     lastItemID,
		VisitorID:      visitorID,
	}

	if user != expectedUser {
		t.Errorf("Expected user %+v, got %+v", expectedUser, user)
	}

	// Update the existing user.
	newLastItemID := "item-2"
	err = driver.Set(userID, newLastItemID, visitorID)
	if err != nil {
		t.Fatalf("Set (update) returned an unexpected error: %v", err)
	}

	// Retrieve the updated user.
	updatedUser, err := driver.Get(userID)
	if err != nil {
		t.Fatalf("Get after update returned an unexpected error: %v", err)
	}

	expectedUpdatedUser := storage.UserRecord{
		UserID:         userID,
		TotalItemsSeen: 2,
		LastItemID:     newLastItemID,
		VisitorID:      visitorID,
	}

	if updatedUser != expectedUpdatedUser {
		t.Errorf("Expected updated user %+v, got %+v", expectedUpdatedUser, updatedUser)
	}
}

// TestConcurrentAccess verifies that KVDriver can handle concurrent Set and Get operations.
func TestConcurrentAccess(t *testing.T) {
	driver := NewKVDriver()
	userID := "concurrentUser"
	visitorID := "visitor-concurrent"

	const numGoroutines = 100
	done := make(chan bool)

	// Concurrently set the user record.
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			lastItemID := fmt.Sprintf("item-%d", i)
			if err := driver.Set(userID, lastItemID, visitorID); err != nil {
				t.Errorf("Set returned an unexpected error: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish.
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Retrieve the user record.
	user, err := driver.Get(userID)
	if err != nil {
		t.Fatalf("Get returned an unexpected error: %v", err)
	}

	// The TotalItemsSeen should be equal to numGoroutines.
	if user.TotalItemsSeen != numGoroutines {
		t.Errorf("Expected TotalItemsSeen to be %d, got %d", numGoroutines, user.TotalItemsSeen)
	}

	// LastItemID should be one of the set values.
	// Since goroutines are concurrent, we can't predict the exact value, but we can check if it starts with the expected prefix.
	if len(user.LastItemID) < len("item-") || user.LastItemID[:len("item-")] != "item-" {
		t.Errorf("LastItemID has unexpected value: %s", user.LastItemID)
	}

	// VisitorID should remain unchanged.
	if user.VisitorID != visitorID {
		t.Errorf("Expected VisitorID to be %s, got %s", visitorID, user.VisitorID)
	}
}

// TestGetNonExistentUser verifies that getting a non-existent user returns the correct error.
func TestGetNonExistentUser(t *testing.T) {
	driver := NewKVDriver()
	nonExistentUserID := "nonexistent123"

	_, err := driver.Get(nonExistentUserID)
	if err == nil {
		t.Fatalf("Expected error when getting non-existent user, got nil")
	}
	if err != errUserNotFound {
		t.Fatalf("Expected errUserNotFound, got %v", err)
	}
}

// TestSetMultipleUsers verifies that multiple users can be set and retrieved independently.
func TestSetMultipleUsers(t *testing.T) {
	driver := NewKVDriver()

	users := []storage.UserRecord{
		{
			UserID:         "user1",
			TotalItemsSeen: 1,
			LastItemID:     "item-1",
			VisitorID:      "visitor-1",
		},
		{
			UserID:         "user2",
			TotalItemsSeen: 2,
			LastItemID:     "item-2",
			VisitorID:      "visitor-2",
		},
		{
			UserID:         "user3",
			TotalItemsSeen: 3,
			LastItemID:     "item-3",
			VisitorID:      "visitor-3",
		},
	}

	// Set each user.
	for _, u := range users {
		for i := 0; i < u.TotalItemsSeen; i++ {
			if err := driver.Set(u.UserID, u.LastItemID, u.VisitorID); err != nil {
				t.Fatalf("Set returned an unexpected error for user %s: %v", u.UserID, err)
			}
		}
	}

	// Retrieve and verify each user.
	for _, expectedUser := range users {
		user, err := driver.Get(expectedUser.UserID)
		if err != nil {
			t.Fatalf("Get returned an unexpected error for user %s: %v", expectedUser.UserID, err)
		}

		if user.UserID != expectedUser.UserID {
			t.Errorf("Expected UserID %s, got %s", expectedUser.UserID, user.UserID)
		}

		if user.TotalItemsSeen != expectedUser.TotalItemsSeen {
			t.Errorf("Expected TotalItemsSeen %d for user %s, got %d", expectedUser.TotalItemsSeen, expectedUser.UserID, user.TotalItemsSeen)
		}

		if user.LastItemID != expectedUser.LastItemID {
			t.Errorf("Expected LastItemID %s for user %s, got %s", expectedUser.LastItemID, expectedUser.UserID, user.LastItemID)
		}

		if user.VisitorID != expectedUser.VisitorID {
			t.Errorf("Expected VisitorID %s for user %s, got %s", expectedUser.VisitorID, expectedUser.UserID, user.VisitorID)
		}
	}
}
