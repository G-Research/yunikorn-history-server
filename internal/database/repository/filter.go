package repository

import (
	"github.com/G-Research/yunikorn-history-server/internal/database/sql"
)

// applyLimitAndOffset adds limit and offset to the sql query.
func applyLimitAndOffset(builder *sql.Builder, limit *int, offset *int) {
	if limit != nil {
		builder.Limit(*limit)
	}
	if offset != nil {
		builder.Offset(*offset)
	}
}
