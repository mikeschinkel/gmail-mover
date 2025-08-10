package sqliteutil

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
)

func NewNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func NewNullInt64(n int64) sql.NullInt64 {
	return sql.NullInt64{Int64: n, Valid: true}
}

func Close(c io.Closer, f func(err error)) {
	f(c.Close())
}

func WarnOnError(err error) {
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
	}
}

//goland:noinspection GoUnusedExportedFunction
func Ptr[T any](a T) *T {
	return &a
}

func EnsureNotNil[T any](v *T, def T) T {
	if v == nil {
		v = &def
	}
	return *v
}

func NewContext() context.Context {
	return context.Background()
}

func panicf(f string, args ...any) {
	panic(fmt.Sprintf(f, args...))
}
