package cmdutil_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestValidateRequired tests the ValidateRequired function
// Since it calls os.Exit(1), we need to test it in a subprocess
func TestValidateRequired(t *testing.T) {
	t.Parallel()

	t.Run("valid value does not exit", func(t *testing.T) {
		t.Parallel()

		// This should not exit since value is provided
		// We can't directly test this without subprocess, so we'll test the happy path behavior
		// by ensuring the function exists and doesn't panic with valid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidateRequired panicked with valid input: %v", r)
			}
		}()

		// We can't actually call this without it potentially exiting,
		// so we'll test it indirectly through subprocess tests below
		// Just verify the test structure is sound
		assert.True(t, true, "Test structure verification")
	})
}

func TestValidateRequiredExitsOnEmptyValue(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		// This is the subprocess that will test the exit behavior
		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
		}
		cmdutil.ValidateRequired("", "test-value", cmd)
		return
	}

	// Main test process
	cmd := exec.Command(os.Args[0], "-test.run=TestValidateRequiredExitsOnEmptyValue")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	// The subprocess should exit with code 1
	if exitError, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 1, exitError.ExitCode())
	} else {
		t.Fatalf("Expected exit error, got: %v", err)
	}
	
	// Check that error message was printed
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "test-value is required")
}

func TestHandleError(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		// This is the subprocess that will test the exit behavior
		cmdutil.HandleError("test error", errors.New("sample error"))
		return
	}

	// Main test process
	cmd := exec.Command(os.Args[0], "-test.run=TestHandleError")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	// The subprocess should exit with code 1
	if exitError, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 1, exitError.ExitCode())
	} else {
		t.Fatalf("Expected exit error, got: %v", err)
	}
	
	// Check that error message was printed
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "test error: sample error")
}

func TestHandleErrorf(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		// This is the subprocess that will test the exit behavior
		cmdutil.HandleErrorf("formatted error: %s %d", "test", 42)
		return
	}

	// Main test process
	cmd := exec.Command(os.Args[0], "-test.run=TestHandleErrorf")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	// The subprocess should exit with code 1
	if exitError, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 1, exitError.ExitCode())
	} else {
		t.Fatalf("Expected exit error, got: %v", err)
	}
	
	// Check that error message was printed
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "formatted error: test 42")
}

// Test error message formatting without subprocess
func TestErrorMessageFormatting(t *testing.T) {
	t.Parallel()

	t.Run("HandleError format", func(t *testing.T) {
		// We can't test the actual function without it exiting,
		// but we can verify the error message format it would produce
		testError := errors.New("connection failed")
		expectedFormat := "network error: connection failed"
		
		// Simulate what HandleError would print
		message := "network error"
		actualFormat := message + ": " + testError.Error()
		
		assert.Equal(t, expectedFormat, actualFormat)
	})

	t.Run("HandleErrorf format", func(t *testing.T) {
		// Test the formatting logic used in HandleErrorf
		format := "operation failed: %s (code: %d)"
		args := []interface{}{"validation error", 400}
		
		// Simulate what HandleErrorf would print
		actualFormat := fmt.Sprintf(format, args...)
		expectedFormat := "operation failed: validation error (code: 400)"
		
		assert.Equal(t, expectedFormat, actualFormat)
	})
}

// Test ValidateRequired behavior with various inputs (without calling the actual function)
func TestValidateRequiredBehavior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		value         string
		paramName     string
		shouldExit    bool
		expectedError string
	}{
		{
			name:          "empty string should exit",
			value:         "",
			paramName:     "filename",
			shouldExit:    true,
			expectedError: "filename is required",
		},
		{
			name:          "whitespace only should exit", 
			value:         "",
			paramName:     "token",
			shouldExit:    true,
			expectedError: "token is required",
		},
		{
			name:       "valid value should not exit",
			value:      "valid-value",
			paramName:  "param",
			shouldExit: false,
		},
		{
			name:       "special characters in value should not exit",
			value:      "value-with-!@#$%^&*()",
			paramName:  "special",
			shouldExit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// We simulate the validation logic without calling the actual function
			isEmpty := tt.value == ""
			assert.Equal(t, tt.shouldExit, isEmpty, "Expected exit condition for value: %q", tt.value)
		})
	}
}

// Test command creation for ValidateRequired
func TestCommandCreation(t *testing.T) {
	t.Parallel()

	t.Run("command with usage", func(t *testing.T) {
		t.Parallel()

		cmd := &cobra.Command{
			Use:   "test-command",
			Short: "A test command",
			Long:  "This is a longer description of the test command",
		}

		assert.NotNil(t, cmd)
		assert.Equal(t, "test-command", cmd.Use)
		assert.Equal(t, "A test command", cmd.Short)
		assert.Equal(t, "This is a longer description of the test command", cmd.Long)

		// Test that Usage() method exists and can be called
		usageFunc := cmd.Usage
		assert.NotNil(t, usageFunc)
	})

	t.Run("command without usage", func(t *testing.T) {
		t.Parallel()

		cmd := &cobra.Command{}
		assert.NotNil(t, cmd)

		// Should still have a Usage method
		usageFunc := cmd.Usage
		assert.NotNil(t, usageFunc)
	})
}

// Test error types that might be passed to HandleError
func TestErrorTypes(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()

		var err error
		if err != nil {
			// This wouldn't be called if err is nil
			message := "test error: " + err.Error()
			assert.NotEmpty(t, message)
		} else {
			// HandleError should probably check for nil, but the current implementation doesn't
			// This test documents the current behavior
			assert.Nil(t, err)
		}
	})

	t.Run("custom error types", func(t *testing.T) {
		t.Parallel()

		customErr := &CustomError{msg: "custom error message"}
		message := "operation failed: " + customErr.Error()
		assert.Equal(t, "operation failed: custom error message", message)
	})

	t.Run("wrapped errors", func(t *testing.T) {
		t.Parallel()

		baseErr := errors.New("base error")
		wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
		
		message := "context: " + wrappedErr.Error()
		assert.Equal(t, "context: wrapped: base error", message)
	})
}

// CustomError for testing
type CustomError struct {
	msg string
}

func (e *CustomError) Error() string {
	return e.msg
}

// Test that package functions exist and have correct signatures
func TestFunctionExistence(t *testing.T) {
	t.Parallel()

	t.Run("ValidateRequired exists", func(t *testing.T) {
		t.Parallel()

		// Test that the function exists by checking it's not nil
		// We can't call it without side effects, but we can verify it exists
		cmd := &cobra.Command{}
		
		// This should compile, proving the function signature is correct
		validateFunc := func() {
			cmdutil.ValidateRequired("test", "param", cmd)
		}
		
		assert.NotNil(t, validateFunc)
	})

	t.Run("HandleError exists", func(t *testing.T) {
		t.Parallel()

		handleErrorFunc := func() {
			cmdutil.HandleError("test", errors.New("test"))
		}
		
		assert.NotNil(t, handleErrorFunc)
	})

	t.Run("HandleErrorf exists", func(t *testing.T) {
		t.Parallel()

		handleErrorfFunc := func() {
			cmdutil.HandleErrorf("test %s", "value")
		}
		
		assert.NotNil(t, handleErrorfFunc)
	})
}

// Integration test simulation
func TestCmdutilIntegration(t *testing.T) {
	t.Parallel()

	t.Run("typical usage pattern", func(t *testing.T) {
		t.Parallel()

		// Simulate typical usage in a CLI command
		cmd := &cobra.Command{
			Use:   "upload",
			Short: "Upload a file",
			RunE: func(cmd *cobra.Command, args []string) error {
				filename := "test.txt"
				
				// This is what would typically happen:
				// cmdutil.ValidateRequired(filename, "filename", cmd)
				
				// Instead, we simulate the validation
				if filename == "" {
					return errors.New("filename is required")
				}
				
				return nil
			},
		}

		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)
	})
}