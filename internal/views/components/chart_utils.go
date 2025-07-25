package components

import "regexp"

// ValidateUpdateInterval validates HTMX update interval format and returns a safe value
func ValidateUpdateInterval(interval string) string {
	// Validate interval format: must be a number followed by s (seconds), m (minutes), or h (hours)
	validInterval := regexp.MustCompile(`^[1-9]\d*[smh]$`)
	
	if validInterval.MatchString(interval) {
		return interval
	}
	
	// Return safe default if validation fails
	return "30s"
}