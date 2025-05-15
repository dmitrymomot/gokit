package sanitizer

var builtInSanitizers = map[string]SanitizeFunc{
	// String sanitizers
	"trim":      trimSanitizer,
	"lower":     lowerSanitizer,
	"upper":     upperSanitizer,
	"replace":   replaceSanitizer,
	"striphtml": stripHTMLSanitizer,
	"escape":    escapeSanitizer,
	"alphanum":  alphanumSanitizer,
	"numeric":   numericSanitizer,
	"truncate":  truncateSanitizer,
	"normalize": normalizeSanitizer,
	"trimspace": trimspaceSanitizer,
	"email":     emailSanitizer,

	// Case sanitizers
	"capitalize": capitalizeSanitizer,
	"camelcase":  camelCaseSanitizer,
	"pascalcase": pascalCaseSanitizer,
	"snakecase":  snakeCaseSanitizer,
	"kebabcase":  kebabCaseSanitizer,
	"ucfirst":    ucfirstSanitizer,
	"lcfirst":    lcfirstSanitizer,

	// Special sanitizers
	"slug": slugSanitizer,
	"uuid": uuidSanitizer,
	"bool": boolSanitizer,
}
