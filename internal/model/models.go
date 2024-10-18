package model

type Metadata struct {
	ID            string `json:"id"`
	CreatedAtNano int64  `json:"createdAtNano"`
	DeletedAtNano *int64 `json:"deletedAtNano,omitempty"`
}
