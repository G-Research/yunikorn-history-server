package model

type Metadata struct {
	CreatedAtNano int64  `json:"createdAtNano"`
	DeletedAtNano *int64 `json:"deletedAtNano,omitempty"`
}
