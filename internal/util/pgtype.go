package util

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func PgTextString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func FormatPgTimestamp(t pgtype.Timestamp) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

func ToPgInt8(v int64) pgtype.Int8 {
	if v == 0 {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: v, Valid: true}
}

func PgInt8Value(t pgtype.Int8) int64 {
	if !t.Valid {
		return 0
	}
	return t.Int64
}
