package pgutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToPgUUID 將 uuid.UUID 轉換為 pgtype.UUID
func UUIDToPgUUIDV5(id uuid.UUID) pgtype.UUID {
	var pgUUID pgtype.UUID
	pgUUID.Bytes = id
	pgUUID.Valid = true
	return pgUUID
}

// PgUUIDToUUID 將 pgtype.UUID 轉換為 uuid.UUID
func PgUUIDToUUIDV5(pgUUID pgtype.UUID) uuid.UUID {
	if !pgUUID.Valid {
		return uuid.Nil
	}
	return uuid.UUID(pgUUID.Bytes)
}

// StringToPgText 將字串指針轉換為 pgtype.Text
func StringToPgTextV5(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// PgTextToString 將 pgtype.Text 轉換為字串指針
func PgTextToStringV5(text pgtype.Text) *string {
	if !text.Valid {
		return nil
	}
	s := text.String
	return &s
}

// TimeToPgTimestamptz 將 time.Time 轉換為 pgtype.Timestamptz
func TimeToPgTimestamptzV5(t *time.Time) pgtype.Timestamptz {
	if t == nil || t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// PgTimestamptzToTime 將 pgtype.Timestamptz 轉換為 time.Time
func PgTimestamptzToTimeV5(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

// BoolToPgBool 將布林值指針轉換為 pgtype.Bool
func BoolToPgBoolV5(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

// PgBoolToBool 將 pgtype.Bool 轉換為布林值指針
func PgBoolToBoolV5(pgBool pgtype.Bool) *bool {
	if !pgBool.Valid {
		return nil
	}
	b := pgBool.Bool
	return &b
}
