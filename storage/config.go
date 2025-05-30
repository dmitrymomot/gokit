package storage

import "time"

// Default constants for configuration
const (
	// DefaultMaxFileSize represents the default maximum allowed file size (10MB)
	DefaultMaxFileSize int64 = 10485760

	// DefaultUploadBasePath is the default base path for uploads
	DefaultUploadBasePath = "uploads"

	// DefaultFieldName is the default form field name for file uploads
	DefaultFieldName = "file"

	// DefaultConnectTimeout is the default connection timeout
	DefaultConnectTimeout = 5 * time.Second

	// DefaultRequestTimeout is the default request timeout
	DefaultRequestTimeout = 60 * time.Second

	// DefaultMaxRetries is the default number of retry attempts
	DefaultMaxRetries = 3

	// DefaultRetryBaseDelay is the default delay between retry attempts
	DefaultRetryBaseDelay = 100 * time.Millisecond
)

// Config holds the configuration required to initialize the S3 client.
type Config struct {
	Key      string `env:"STORAGE_KEY,required"`
	Secret   string `env:"STORAGE_SECRET,required"`
	Region   string `env:"STORAGE_REGION,required"`
	Bucket   string `env:"STORAGE_BUCKET,required"`
	Endpoint string `env:"STORAGE_ENDPOINT,required"`
	CDN      string `env:"STORAGE_CDN"`

	MaxFileSize    int64  `env:"STORAGE_MAX_FILE_SIZE" envDefault:"10485760"` // 10MB
	UploadBasePath string `env:"STORAGE_BASE_PATH" envDefault:"uploads"`
	ForcePathStyle bool   `env:"STORAGE_FORCE_PATH_STYLE" envDefault:"false"`

	ConnectTimeout time.Duration `env:"STORAGE_CONNECT_TIMEOUT" envDefault:"5s"`
	RequestTimeout time.Duration `env:"STORAGE_REQUEST_TIMEOUT" envDefault:"60s"`

	MaxRetries     int           `env:"STORAGE_MAX_RETRIES" envDefault:"3"`
	RetryBaseDelay time.Duration `env:"STORAGE_RETRY_BASE_DELAY" envDefault:"100ms"`
}
