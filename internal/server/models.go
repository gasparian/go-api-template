package server

// GetRequestModel ...
type GetRequestModel struct {
	UserID string `json:"userId"`
}

// GetResponseModel ...
type GetResponseModel struct {
	UserID         string `json:"userId"`
	TotalItemsSeen int    `json:"totalItemsSeen"`
	LastItemID     string `json:"lastItemId,omitempty"`
	VisitorID      string `json:"visitorId,omitempty"`
}

// PostRequestModel ...
type PostRequestModel struct {
	UserID string `json:"userId"`
	ItemID string `json:"itemId"`
}

// PostResponseModel ...
type PostResponseModel struct {
	UserID         string `json:"userId"`
	TotalItemsSeen int    `json:"totalItemsSeen"`
}
