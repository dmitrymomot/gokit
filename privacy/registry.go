package privacy

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// MaskingRegistry implements the Registry interface for managing maskers.
type MaskingRegistry struct {
	// mu protects access to the maskers map
	mu sync.RWMutex

	// maskers is a map of category to masker.
	maskers map[DataCategory]Masker

	// defaultMasker is used when no specific masker is registered for a category.
	defaultMasker Masker
}

// NewMaskingRegistry creates a new MaskingRegistry with optional default masker.
func NewMaskingRegistry(defaultMasker Masker) *MaskingRegistry {
	return &MaskingRegistry{
		maskers:       make(map[DataCategory]Masker),
		defaultMasker: defaultMasker,
	}
}

// RegisterMasker registers a masker for a specific data category.
func (r *MaskingRegistry) RegisterMasker(category DataCategory, masker Masker) error {
	if masker == nil {
		return errors.Join(ErrInvalidMasker, ErrNilMasker)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.maskers[category] = masker
	return nil
}

// UnregisterMasker removes a masker for a specific data category.
func (r *MaskingRegistry) UnregisterMasker(category DataCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.maskers[category]; !exists {
		return errors.Join(ErrMaskerNotFound, fmt.Errorf("no masker registered for category: %s", category))
	}

	delete(r.maskers, category)
	return nil
}

// GetMasker retrieves a masker for a specific data category.
func (r *MaskingRegistry) GetMasker(category DataCategory) (Masker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	masker, ok := r.maskers[category]
	if !ok {
		if r.defaultMasker != nil {
			return r.defaultMasker, nil
		}
		return nil, errors.Join(ErrMaskerNotFound, fmt.Errorf("no masker registered for category: %s", category))
	}

	return masker, nil
}

// MaskByCategory masks data according to its category.
func (r *MaskingRegistry) MaskByCategory(ctx context.Context, category DataCategory, data any) (any, error) {
	masker, err := r.GetMasker(category)
	if err != nil {
		return nil, err
	}

	if !masker.CanMask(data) {
		return nil, errors.Join(ErrUnsupportedType, fmt.Errorf("masker cannot handle data type for category: %s", category))
	}

	return masker.Mask(ctx, data)
}

// DetectionRule is a function that detects if a piece of data belongs to a specific category.
type DetectionRule func(data any) bool

// AutoMaskingRegistry extends MaskingRegistry with automatic category detection.
type AutoMaskingRegistry struct {
	*MaskingRegistry

	// detectionRules maps a category to a detection rule.
	detectionRules map[DataCategory]DetectionRule
}

// NewAutoMaskingRegistry creates a new AutoMaskingRegistry with optional default masker.
func NewAutoMaskingRegistry(defaultMasker Masker) *AutoMaskingRegistry {
	return &AutoMaskingRegistry{
		MaskingRegistry: NewMaskingRegistry(defaultMasker),
		detectionRules:  make(map[DataCategory]DetectionRule),
	}
}

// RegisterDetectionRule registers a detection rule for a specific data category.
func (r *AutoMaskingRegistry) RegisterDetectionRule(category DataCategory, rule DetectionRule) error {
	if rule == nil {
		return errors.Join(ErrInvalidRule, ErrNilDetectionRule)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.detectionRules[category] = rule
	return nil
}

// UnregisterDetectionRule removes a detection rule for a specific data category.
func (r *AutoMaskingRegistry) UnregisterDetectionRule(category DataCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.detectionRules[category]; !exists {
		return errors.Join(ErrRuleNotFound, fmt.Errorf("no detection rule registered for category: %s", category))
	}

	delete(r.detectionRules, category)
	return nil
}

// DetectCategory attempts to detect the category of the given data.
func (r *AutoMaskingRegistry) DetectCategory(data any) (DataCategory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for category, rule := range r.detectionRules {
		if rule(data) {
			return category, true
		}
	}

	return "", false
}

// AutoMask automatically detects the category of the data and applies the appropriate masker.
func (r *AutoMaskingRegistry) AutoMask(ctx context.Context, data any) (any, error) {
	category, detected := r.DetectCategory(data)
	if !detected {
		if r.defaultMasker != nil && r.defaultMasker.CanMask(data) {
			return r.defaultMasker.Mask(ctx, data)
		}
		return nil, errors.Join(ErrCategoryNotDetected, fmt.Errorf("could not detect category for data"))
	}

	return r.MaskByCategory(ctx, category, data)
}
