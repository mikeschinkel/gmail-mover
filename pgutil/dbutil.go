package pgutil

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func ExecSQLFile(ctx context.Context, filename string) (err error) {
	var sql []byte

	// Read test sql SQL file
	sql, err = os.ReadFile(filename)
	if err != nil {
		goto end
	}

	// Execute the sql
	err = ExecSQL(ctx, string(sql))
	if err != nil {
		goto end
	}
end:
	return err
}

func ExecSQL(ctx context.Context, sql string) (err error) {
	_, err = db.Execute(ctx, sql)
	return err
}

func InitializeDB(ctx context.Context, connStr string, dbName ...string) (err error) {
	var conn *pgx.Conn
	if len(dbName) != 0 {
		connStr, err = changePostgresDBName(connStr, dbName[0])
	}
	if err != nil {
		goto end
	}
	conn, err = pgx.Connect(ctx, connStr)
	if err != nil {
		goto end
	}
	SetDB(NewDB(conn))
end:
	return err
}

var db *DB

func GetDB() *DB {
	if db == nil {
		panic("DB is not initialized")
	}
	return db
}

func SetDB(newDB *DB) {
	db = newDB
}

type Querier interface {
	// Define your query interface methods here based on sqlc generated code
}

type DB struct {
	Querier
	conn *pgx.Conn // Keep a reference to the raw DB for any functions not yet migrated
}

func (db *DB) Close(ctx context.Context) (err error) {
	return db.conn.Close(ctx)
}

func (db *DB) Execute(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.conn.Exec(ctx, query, args...)
}

// NewDB creates a new DB
func NewDB(conn *pgx.Conn) (db *DB) {
	db = &DB{
		// Querier: sqlc.New(conn), // Replace with actual sqlc.New call when implementing
		conn: conn,
	}
	runtime.SetFinalizer(db, func(db *DB) {
		err := db.conn.Close(context.Background())
		if err != nil {
			slog.Warn("failed to close database", "error", err.Error(), "type_map", db.conn.TypeMap())
		}
	})
	return db
}

// Query sends a query to the server and returns a list of rows as pgx.Rows to
// read the results. Only errors encountered sending the query and initializing
// Rows will be returned. Err() on the returned Rows must be checked after the
// Rows is closed to determine if the query executed successfully.
//
// The returned Rows must be closed before the connection can be used again. It
// is safe to attempt to read from the returned Rows even if an error is
// returned. The error will be the available in rows.Err() after rows are closed.
// It is allowed to ignore the error returned from Query and handle it in Rows.
//
// It is possible for a call of FieldDescriptions on the returned Rows to return
// nil even if the Query call did not return an error.
//
// It is possible for a query to return one or more rows before encountering an
// error. In most cases the rows should be collected before processing rather
// than processed while receiving each row. This avoids the possibility of the
// application processing rows from a query that the server rejected. The
// CollectRows function is useful here.
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.conn.Query(ctx, query, args...)
}

// QueryRow is a convenience wrapper over Query. Any error that occurs while
// querying is deferred until calling Scan on the returned Row. That pgx.Row will
// error with pgx.ErrNoRows if no rows are returned.
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.conn.QueryRow(ctx, query, args...)
}

var postgresConnStrRegex = regexp.MustCompile(`^(postgres://[^/]+)/(.+)(\?.+)*$`)

func changePostgresDBName(connStr, dbName string) (_ string, err error) {
	// Parse the URL

	matches := postgresConnStrRegex.FindStringSubmatch(connStr)
	if matches == nil {
		err = fmt.Errorf("invalid Postgres connection string: '%s'", connStr)
		goto end
	}
	connStr = fmt.Sprintf("%s/%s%s", matches[1], dbName, matches[3])

end:
	return connStr, err
}

//goland:noinspection GoUnusedFunction
func extractPostgresDBName(connStr string) (dbName string) {
	matches := postgresConnStrRegex.FindStringSubmatch(connStr)
	if matches == nil {
		dbName = "postgres" // The default DB
		goto end
	}
	dbName = matches[2]

end:
	return connStr
}
