package kubernetes

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
)

// SanitizeError sanitizes Kubernetes errors to remove sensitive information
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	// Handle Kubernetes API errors
	if k8sErr, ok := err.(errors.APIStatus); ok {
		status := k8sErr.Status()

		// Create a sanitized error message
		msg := fmt.Sprintf("Kubernetes API error: %s", status.Message)

		// Add reason if available
		if status.Reason != "" {
			msg = fmt.Sprintf("%s (reason: %s)", msg, status.Reason)
		}

		// Add code if available
		if status.Code != 0 {
			msg = fmt.Sprintf("%s (code: %d)", msg, status.Code)
		}

		return fmt.Errorf("%s", msg)
	}

	// For other errors, just return the error message without stack traces
	errStr := err.Error()

	// Remove potential sensitive paths
	errStr = strings.ReplaceAll(errStr, os.Getenv("HOME"), "$HOME")

	return fmt.Errorf("%s", errStr)
}

// IsNotFound checks if error is a "not found" error
func IsNotFound(err error) bool {
	return errors.IsNotFound(err)
}

// IsForbidden checks if error is a "forbidden" (RBAC) error
func IsForbidden(err error) bool {
	return errors.IsForbidden(err)
}

// IsUnauthorized checks if error is an "unauthorized" error
func IsUnauthorized(err error) bool {
	return errors.IsUnauthorized(err)
}
