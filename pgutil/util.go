package pgutil

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func NewPGText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}
func NewPGBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}
func NewPGTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time: t,
	}
}
