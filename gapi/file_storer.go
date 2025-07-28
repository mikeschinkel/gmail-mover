package gapi

// FileStorer provides file operations for Gmail API
type FileStorer interface {
	Load(filename string, data any) error
	Save(filename string, data any) error
	Exists(filename string) bool
}
