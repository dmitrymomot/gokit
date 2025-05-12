package validator

import "maps"

// Option defines a function type for configuring a Validator
type Option func(*Validator) error

// WithErrorTranslator sets a custom error translator
func WithErrorTranslator(translator ErrorTranslatorFunc) Option {
	return func(v *Validator) error {
		if translator != nil {
			v.errorTranslator = translator
		} else {
			v.errorTranslator = defaultErrorTranslator
		}
		return nil
	}
}

// WithSeparators sets custom separators for the validator
func WithSeparators(ruleSep, paramSep, paramListSep string) Option {
	return func(v *Validator) error {
		// Validate separators
		if ruleSep == "" || paramSep == "" || paramListSep == "" {
			return ErrInvalidSeparatorConfiguration
		}
		if len(ruleSep) != 1 || len(paramSep) != 1 || len(paramListSep) != 1 {
			return ErrInvalidSeparatorConfiguration
		}
		if ruleSep == paramSep || ruleSep == paramListSep || paramSep == paramListSep {
			return ErrInvalidSeparatorConfiguration
		}

		v.ruleSeparator = ruleSep
		v.paramSeparator = paramSep
		v.paramListSeparator = paramListSep
		return nil
	}
}

// WithValidators adds specific validators to the Validator instance
func WithValidators(validatorNames ...string) Option {
	return func(v *Validator) error {
		// Clear existing validators to ensure only specified ones are loaded
		v.validatorsMutex.Lock()
		v.validators = make(map[string]ValidationFunc)
		v.validatorsMutex.Unlock()

		for _, name := range validatorNames {
			if fn, exists := builtInValidators[name]; exists {
				v.validatorsMutex.Lock()
				v.validators[name] = fn
				v.validatorsMutex.Unlock()
			}
		}
		return nil
	}
}

// WithAllValidators adds all built-in validators to the Validator instance
func WithAllValidators() Option {
	return func(v *Validator) error {
		v.validatorsMutex.Lock()
		defer v.validatorsMutex.Unlock()
		maps.Copy(v.validators, builtInValidators)
		return nil
	}
}

// WithExcept adds all built-in validators except specified ones
func WithExcept(excludedNames ...string) Option {
	return func(v *Validator) error {
		v.validatorsMutex.Lock()
		defer v.validatorsMutex.Unlock()
		for _, name := range excludedNames {
			delete(v.validators, name)
		}
		return nil
	}
}

// WithCustomValidator adds a custom validator function
func WithCustomValidator(name string, fn ValidationFunc) Option {
	return func(v *Validator) error {
		if name == "" || fn == nil {
			return ErrInvalidValidatorConfiguration
		}
		v.validatorsMutex.Lock()
		v.validators[name] = fn
		v.validatorsMutex.Unlock()
		return nil
	}
}

// WithFieldNameTag sets the tag name used for identifying field validation rules
func WithFieldNameTag(tagName string) Option {
	return func(v *Validator) error {
		if tagName == "" {
			return ErrInvalidValidatorConfiguration
		}
		v.fieldNameTag = tagName
		return nil
	}
}
