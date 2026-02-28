package storage

// UserRecord ...
type UserRecord struct {
	UserID         string `json:"userId"`
	TotalItemsSeen int    `json:"totalItemsSeen"`
	LastItemID     string `json:"lastItemId,omitempty"`
	VisitorID      string `json:"visitorId,omitempty"`
}
