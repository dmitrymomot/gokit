package sanitizer

// Option configures a Sanitizer instance.
type Option func(*Sanitizer) error

// WithRuleSeparator sets the rule separator for the sanitizer.
// Default separator is ";".
func WithRuleSeparator(separator string) Option {
	return func(s *Sanitizer) error {
		if separator == "" {
			return ErrInvalidSanitizerConfiguration
		}
		s.ruleSeparator = separator
		return nil
	}
}

// WithParamSeparator sets the parameter separator for the sanitizer.
// Default separator is ":".
func WithParamSeparator(separator string) Option {
	return func(s *Sanitizer) error {
		if separator == "" {
			return ErrInvalidSanitizerConfiguration
		}
		s.paramSeparator = separator
		return nil
	}
}

// WithParamListSeparator sets the parameter list separator for the sanitizer.
// Default separator is ",".
func WithParamListSeparator(separator string) Option {
	return func(s *Sanitizer) error {
		if separator == "" {
			return ErrInvalidSanitizerConfiguration
		}
		s.paramListSeparator = separator
		return nil
	}
}

// WithFieldNameTag sets the field name tag for the sanitizer.
// Default is empty, which uses the struct field name.
func WithFieldNameTag(tag string) Option {
	return func(s *Sanitizer) error {
		s.fieldNameTag = tag
		return nil
	}
}

// WithSanitizers sets the custom sanitizers map for the sanitizer.
func WithSanitizers(sanitizers map[string]SanitizeFunc) Option {
	return func(s *Sanitizer) error {
		if sanitizers == nil {
			return ErrInvalidSanitizerConfiguration
		}
		s.sanitizersMu.Lock()
		defer s.sanitizersMu.Unlock()
		
		for name, fn := range sanitizers {
			if name == "" || fn == nil {
				return ErrInvalidSanitizerConfiguration
			}
			s.sanitizers[name] = fn
		}
		return nil
	}
}
