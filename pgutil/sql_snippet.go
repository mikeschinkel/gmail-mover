package pgutil

type SQLSnippet struct {
	Message string
	Query   string
	Error   error
}
