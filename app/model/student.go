package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Validasi deklaratif: aturan dekat dengan data
type CreateStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,min=5,max=20,nim"`
	Name     string  `json:"name" validate:"required,min=3,max=100"`
	Grade    float64 `json:"grade" validate:"required,gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}

type ReplaceStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,min=5,max=20,nim"`
	Name     string  `json:"name" validate:"required,min=3,max=100"`
	Grade    float64 `json:"grade" validate:"required,gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}

// PATCH: pointer + omitnil agar "tidak dikirim" bisa dibedakan dari "dikirim kosong"
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty" validate:"omitnil,min=5,max=20,nim"`
	Name     *string  `json:"name,omitempty" validate:"omitnil,min=3,max=100"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,gte=0,lte=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
