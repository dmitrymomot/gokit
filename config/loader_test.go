package config_test

import (
	"errors"
	"os"
	"testing"

	"github.com/dmitrymomot/gokit/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration structs
type TestConfigDefault struct {
	TestString string `env:"TEST_STRING_DEFAULT" envDefault:"default_value"`
	TestInt    int    `env:"TEST_INT_DEFAULT" envDefault:"42"`
	TestBool   bool   `env:"TEST_BOOL_DEFAULT" envDefault:"true"`
}

type TestConfigSuccess struct {
	TestString string `env:"TEST_STRING_SUCCESS" envDefault:"default_value"`
	TestInt    int    `env:"TEST_INT_SUCCESS" envDefault:"42"`
	TestBool   bool   `env:"TEST_BOOL_SUCCESS" envDefault:"true"`
}

type TestConfigSingleton struct {
	TestString string `env:"TEST_STRING_SINGLETON" envDefault:"default_value"`
}

type TestConfigDifferent1 struct {
	Value string `env:"VALUE_TYPE1" envDefault:"default1"`
}

type TestConfigDifferent2 struct {
	Value string `env:"VALUE_TYPE2" envDefault:"default2"`
}

type RequiredConfig struct {
	Required string `env:"REQUIRED_VALUE,required"`
}

func TestLoad_Success(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("TEST_STRING_SUCCESS", "test_value")
	t.Setenv("TEST_INT_SUCCESS", "100")
	t.Setenv("TEST_BOOL_SUCCESS", "false")

	// Load configuration
	cfg, err := config.Load[TestConfigSuccess]()

	// Assert no error
	require.NoError(t, err, "Load should not return an error with valid environment variables")

	// Assert values are correctly loaded
	assert.Equal(t, "test_value", cfg.TestString, "TestString should match environment variable")
	assert.Equal(t, 100, cfg.TestInt, "TestInt should match environment variable")
	assert.Equal(t, false, cfg.TestBool, "TestBool should match environment variable")
}

func TestLoad_DefaultValues(t *testing.T) {
	// Clean environment variables to ensure defaults are used
	os.Unsetenv("TEST_STRING_DEFAULT")
	os.Unsetenv("TEST_INT_DEFAULT")
	os.Unsetenv("TEST_BOOL_DEFAULT")

	// Load configuration
	cfg, err := config.Load[TestConfigDefault]()

	// Assert no error
	require.NoError(t, err, "Load should not return an error when using default values")

	// Assert default values are correctly loaded
	assert.Equal(t, "default_value", cfg.TestString, "TestString should use default value")
	assert.Equal(t, 42, cfg.TestInt, "TestInt should use default value")
	assert.Equal(t, true, cfg.TestBool, "TestBool should use default value")
}

func TestLoad_MissingRequired(t *testing.T) {
	// Ensure REQUIRED_VALUE is not set
	os.Unsetenv("REQUIRED_VALUE")

	// Load configuration
	_, err := config.Load[RequiredConfig]()

	// Assert error is returned
	require.Error(t, err, "Load should return an error when a required value is missing")
	assert.True(t, errors.Is(err, config.ErrParsingConfig), "Error should be ErrParsingConfig")
}

func TestLoad_Singleton(t *testing.T) {
	// Set environment variables
	t.Setenv("TEST_STRING_SINGLETON", "first_value")
	
	// First load
	firstConfig, err := config.Load[TestConfigSingleton]()
	require.NoError(t, err, "First load should not return an error")
	
	// Change environment variable
	t.Setenv("TEST_STRING_SINGLETON", "second_value")
	
	// Second load - should return cached version, not new value
	secondConfig, err := config.Load[TestConfigSingleton]()
	require.NoError(t, err, "Second load should not return an error")
	
	// Assert both configs have the same value (the first one)
	assert.Equal(t, firstConfig.TestString, secondConfig.TestString, 
		"Both configs should have the same value due to singleton pattern")
	assert.Equal(t, "first_value", secondConfig.TestString, 
		"Second config should have the first value due to caching")
}

func TestLoad_DifferentTypes(t *testing.T) {
	// Set environment variables
	t.Setenv("VALUE_TYPE1", "test_type1")
	t.Setenv("VALUE_TYPE2", "test_type2")
	
	// Load first config type
	config1, err := config.Load[TestConfigDifferent1]()
	require.NoError(t, err, "Loading first config type should not error")
	
	// Load second config type
	config2, err := config.Load[TestConfigDifferent2]()
	require.NoError(t, err, "Loading second config type should not error")
	
	// Assert each has the correct value
	assert.Equal(t, "test_type1", config1.Value, "First config should have its own value")
	assert.Equal(t, "test_type2", config2.Value, "Second config should have its own value")
}
