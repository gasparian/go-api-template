package storage

// Storage abstraction over the storage driver
type Storage interface {
	Ping() error
	Set(userID, lastItemID, visitorID string) error
	Get(userId string) (UserRecord, error)
}
