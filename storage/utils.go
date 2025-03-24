package storage

import "strings"

// Map of content types to file extensions for O(1) lookup
var contentTypeMap = map[string]string{
	// Images
	"image/jpeg":               ".jpg",
	"image/jpg":                ".jpg",
	"image/png":                ".png",
	"image/gif":                ".gif",
	"image/bmp":                ".bmp",
	"image/webp":               ".webp",
	"image/tiff":               ".tiff",
	"image/svg+xml":            ".svg",
	"image/x-icon":             ".ico",
	"image/vnd.microsoft.icon": ".ico",
	"image/heic":               ".heic",
	"image/heif":               ".heif",
	"image/avif":               ".avif",
	"image/x-ms-bmp":           ".bmp",

	// Documents
	"application/pdf":    ".pdf",
	"application/msword": ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"application/vnd.ms-excel": ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
	"application/vnd.ms-powerpoint":                                             ".ppt",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
	"text/plain":      ".txt",
	"text/csv":        ".csv",
	"text/html":       ".html",
	"text/xml":        ".xml",
	"application/rtf": ".rtf",
	"application/vnd.oasis.opendocument.text":         ".odt",
	"application/vnd.oasis.opendocument.spreadsheet":  ".ods",
	"application/vnd.oasis.opendocument.presentation": ".odp",
	"application/vnd.apple.pages":                     ".pages",
	"application/vnd.apple.numbers":                   ".numbers",
	"application/vnd.apple.keynote":                   ".key",
	"application/epub+zip":                            ".epub",
	"application/xhtml+xml":                           ".xhtml",
	"application/xml":                                 ".xml",
	"application/vnd.visio":                           ".vsd",
	"application/vnd.ms-xpsdocument":                  ".xps",

	// Audio
	"audio/mpeg":       ".mp3",
	"audio/wav":        ".wav",
	"audio/x-wav":      ".wav",
	"audio/ogg":        ".ogg",
	"audio/aac":        ".aac",
	"audio/flac":       ".flac",
	"audio/midi":       ".midi",
	"audio/x-midi":     ".midi",
	"audio/webm":       ".weba",
	"audio/opus":       ".opus",
	"audio/x-m4a":      ".m4a",
	"audio/amr":        ".amr",
	"audio/mp4":        ".m4a",
	"audio/x-matroska": ".mka",
	"audio/vnd.wave":   ".wav",

	// Video
	"video/mp4":        ".mp4",
	"video/mpeg":       ".mpeg",
	"video/quicktime":  ".mov",
	"video/x-msvideo":  ".avi",
	"video/webm":       ".webm",
	"video/x-matroska": ".mkv",
	"video/x-flv":      ".flv",
	"video/3gpp":       ".3gp",
	"video/3gpp2":      ".3g2",
	"video/x-ms-wmv":   ".wmv",
	"video/ogg":        ".ogv",
	"video/x-m4v":      ".m4v",
	"video/mp2t":       ".ts",
	"video/x-ms-asf":   ".asf",

	// Archives
	"application/zip":                       ".zip",
	"application/x-rar-compressed":          ".rar",
	"application/x-tar":                     ".tar",
	"application/gzip":                      ".gz",
	"application/x-7z-compressed":           ".7z",
	"application/x-bzip":                    ".bz",
	"application/x-bzip2":                   ".bz2",
	"application/vnd.rar":                   ".rar",
	"application/x-zip-compressed":          ".zip",
	"application/x-gzip":                    ".gz",
	"application/x-lzma":                    ".lzma",
	"application/x-xz":                      ".xz",
	"application/vnd.debian.binary-package": ".deb",
	"application/x-rpm":                     ".rpm",
	"application/x-stuffit":                 ".sit",
	"application/x-apple-diskimage":         ".dmg",
	"application/x-iso9660-image":           ".iso",

	// Programming
	"application/json":         ".json",
	"application/javascript":   ".js",
	"text/css":                 ".css",
	"application/x-httpd-php":  ".php",
	"application/x-sh":         ".sh",
	"text/x-python":            ".py",
	"text/x-java-source":       ".java",
	"text/x-go":                ".go",
	"text/x-typescript":        ".ts",
	"text/markdown":            ".md",
	"application/x-ruby":       ".rb",
	"text/x-csrc":              ".c",
	"text/x-chdr":              ".h",
	"text/x-c++src":            ".cpp",
	"text/x-c++hdr":            ".hpp",
	"text/x-csharp":            ".cs",
	"text/x-perl":              ".pl",
	"application/x-powershell": ".ps1",
	"application/x-yaml":       ".yaml",
	"text/yaml":                ".yaml",
	"application/toml":         ".toml",
	"application/wasm":         ".wasm",
	"application/dart":         ".dart",
	"text/jsx":                 ".jsx",
	"text/tsx":                 ".tsx",

	// Fonts
	"font/ttf":                      ".ttf",
	"font/otf":                      ".otf",
	"font/woff":                     ".woff",
	"font/woff2":                    ".woff2",
	"application/vnd.ms-fontobject": ".eot",
	"font/collection":               ".ttc",

	// Others
	"application/vnd.android.package-archive":           ".apk",
	"application/x-executable":                          ".exe",
	"application/octet-stream":                          ".bin",
	"application/x-shockwave-flash":                     ".swf",
	"application/vnd.apple.installer+xml":               ".mpkg",
	"application/vnd.mozilla.xul+xml":                   ".xul",
	"application/x-msdownload":                          ".exe",
	"application/vnd.microsoft.portable-executable":     ".exe",
	"application/x-msdos-program":                       ".exe",
	"application/x-ms-shortcut":                         ".lnk",
	"application/x-sqlite3":                             ".sqlite",
	"application/x-deb":                                 ".deb",
	"application/vnd.flatpak":                           ".flatpak",
	"application/vnd.appimage":                          ".appimage",
	"application/vnd.docker.image.rootfs.diff.tar.gzip": ".tar.gz",
	"application/pkcs12":                                ".p12",
	"application/pkcs8":                                 ".p8",
	"application/pgp-encrypted":                         ".pgp",
	"application/x-pkcs7-certificates":                  ".p7b",
	"application/sql":                                   ".sql",
}

// getExtByContentType returns the file extension based on the content type.
// If the content type is unknown, it returns an empty string.
func getExtByContentType(contentType string) string {
	// Extract just the MIME type without parameters
	// Handle cases like "text/plain; charset=utf-8" by taking only "text/plain"
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	// Look up the extension from the map
	if ext, ok := contentTypeMap[contentType]; ok {
		return ext
	}
	return ""
}
