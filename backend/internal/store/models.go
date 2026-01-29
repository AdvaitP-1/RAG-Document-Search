package store

import "time"

type Collection struct {
	ID          string    `json:"id"`
	OwnerUserID string    `json:"owner_user_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
}

type Document struct {
	ID           string    `json:"id"`
	CollectionID string    `json:"collection_id"`
	OwnerUserID  string    `json:"owner_user_id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type IngestionJob struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	Error      *string   `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
