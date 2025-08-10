package sqliteutil

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mikeschinkel/gmover/sqlcx"
)

var _ sqlcx.DataStore = (*SqliteDataStore)(nil)

const macOSConfigSubdir = "/Library/Application Support"
const desiredConfigSubdir = ".config"

type SqliteDataStore struct {
	path    string
	AppName string
	db      *sql.DB
	ddlFunc func() string
}

type Args struct {
	AppName string
	DDLFunc func() string
}

func (ds *SqliteDataStore) Filepath() string {
	return filepath.Join(ds.path, ds.AppName+".db")
}

func NewSqliteDataStore(args Args) sqlcx.DataStore {
	return &SqliteDataStore{
		AppName: args.AppName,
		ddlFunc: args.DDLFunc,
	}
}

func (ds *SqliteDataStore) Initialize(ctx context.Context) (err error) {
	var configDir string

	configDir, err = os.UserConfigDir()
	if err != nil {
		err = ErrFailedToGetConfigPath
		goto end
	}
	// Move macOS config dir to be ~/.config vs. ~/Library/Application Support
	if strings.HasSuffix(configDir, macOSConfigSubdir) {
		configDir = filepath.Join(
			configDir[:len(configDir)-len(macOSConfigSubdir)],
			desiredConfigSubdir,
			ds.AppName,
		)
	}
	ds.path = configDir

	slog.Info("Initializing data store",
		"data_store", relativeToHomeDir(ds.Filepath()),
	)

	err = ds.Open()
	if err != nil {
		goto end
	}
	if ds.ddlFunc != nil {
		err = ds.Query(ctx, ds.ddlFunc())
		if err != nil {
			goto end
		}
	}
end:
	return err
}

func (ds *SqliteDataStore) Open() (err error) {
	err = os.MkdirAll(filepath.Dir(ds.Filepath()), os.ModePerm)
	if err != nil {
		goto end
	}
	ds.db, err = sql.Open("sqlite3", ds.Filepath())
end:
	return err
}

func (ds *SqliteDataStore) Query(ctx context.Context, sql string) (err error) {
	_, err = ds.db.ExecContext(ctx, sql)
	return err
}

func (ds *SqliteDataStore) DB() sqlcx.DBTX {
	return ds.db
}

func relativeToHomeDir(fp string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to get user home directory")
	}
	rel, err := filepath.Rel(home, fp)
	if err != nil {
		panicf("Unable to get relative path to %s", fp)
	}
	return "~/" + rel
}
