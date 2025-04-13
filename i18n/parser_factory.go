package i18n

import "strings"

// NewParserForFile returns a parser based on the file extension
func NewParserForFile(filename string) Parser {
	ext := getFileExtension(filename)
	
	switch strings.ToLower(ext) {
	case "json":
		return NewJSONParser()
	case "yaml", "yml":
		return NewYAMLParser()
	default:
		return nil
	}
}

// getFileExtension extracts the extension from a filename
func getFileExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[idx+1:]
	}
	return ""
}