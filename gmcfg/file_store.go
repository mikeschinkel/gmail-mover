package gmcfg

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// FileStore handles configuration directory and file operations
type FileStore struct {
	appName string
	baseDir string
}

// NewFileStore creates a new file store
func NewFileStore(appName string) *FileStore {
	return &FileStore{
		appName: appName,
	}
}

// ConfigDir returns the configuration directory path
func (fs *FileStore) ConfigDir() (dir string, err error) {
	var homeDir string

	if fs.baseDir != "" {
		dir = fs.baseDir
		goto end
	}

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	dir = filepath.Join(homeDir, ".config", fs.appName)

end:
	return dir, err
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func (fs *FileStore) EnsureConfigDir() (err error) {
	var dir string

	dir, err = fs.ConfigDir()
	if err != nil {
		goto end
	}

	err = os.MkdirAll(dir, 0755)

end:
	return err
}

// filepath returns the full path to a configuration file
func (fs *FileStore) filepath(filename string) (path string, err error) {
	var dir string

	dir, err = fs.ConfigDir()
	if err != nil {
		goto end
	}

	path = filepath.Join(dir, filename)

end:
	return path, err
}

// Save writes data as JSON to a config file
func (fs *FileStore) Save(filename string, data any) (err error) {
	var path string
	var jsonData []byte
	var file *os.File

	ensureLogger()

	err = fs.EnsureConfigDir()
	if err != nil {
		goto end
	}

	path, err = fs.filepath(filename)
	if err != nil {
		goto end
	}

	jsonData, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		goto end
	}

	file, err = os.Create(path)
	if err != nil {
		goto end
	}
	defer mustClose(file)

	_, err = file.Write(jsonData)

end:
	return err
}

// Load reads JSON data from a config file
func (fs *FileStore) Load(filename string, data any) (err error) {
	var path string
	var jsonData []byte

	path, err = fs.filepath(filename)
	if err != nil {
		goto end
	}

	jsonData, err = os.ReadFile(path)
	if err != nil {
		goto end
	}

	err = json.Unmarshal(jsonData, data)

end:
	return err
}

// Append appends content to a file
func (fs *FileStore) Append(filename string, content []byte) (err error) {
	var path string
	var file *os.File

	err = fs.EnsureConfigDir()
	if err != nil {
		goto end
	}

	path, err = fs.filepath(filename)
	if err != nil {
		goto end
	}

	file, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		goto end
	}
	defer mustClose(file)

	_, err = file.Write(content)
	if err != nil {
		goto end
	}

	err = file.Sync()

end:
	return err
}

// Exists checks if a configuration file exists
func (fs *FileStore) Exists(filename string) bool {
	var path string
	var err error

	path, err = fs.filepath(filename)
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// SetBaseDir sets a custom base directory for testing
func (fs *FileStore) SetBaseDir(dir string) {
	fs.baseDir = dir
}
