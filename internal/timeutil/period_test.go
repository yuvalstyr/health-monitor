package timeutil

import (
	"testing"
	"time"
)

func TestCalculateCurrentPeriodStart_Weekly(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{
			name:     "Sunday should return same Sunday",
			date:     "2024-01-14", // Sunday
			expected: "2024-01-14",
		},
		{
			name:     "Monday should return previous Sunday",
			date:     "2024-01-15", // Monday
			expected: "2024-01-14", // Sunday
		},
		{
			name:     "Tuesday should return previous Sunday",
			date:     "2024-01-16", // Tuesday
			expected: "2024-01-14", // Sunday
		},
		{
			name:     "Wednesday should return previous Sunday",
			date:     "2024-01-17", // Wednesday
			expected: "2024-01-14", // Sunday
		},
		{
			name:     "Thursday should return previous Sunday",
			date:     "2024-01-18", // Thursday
			expected: "2024-01-14", // Sunday
		},
		{
			name:     "Friday should return previous Sunday",
			date:     "2024-01-19", // Friday
			expected: "2024-01-14", // Sunday
		},
		{
			name:     "Saturday should return previous Sunday",
			date:     "2024-01-20", // Saturday
			expected: "2024-01-14", // Sunday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			
			result := CalculateCurrentPeriodStart("weekly", date)
			
			if !result.Equal(expected) {
				t.Errorf("CalculateCurrentPeriodStart() = %v, want %v", result.Format("2006-01-02"), expected.Format("2006-01-02"))
			}
		})
	}
}

func TestCalculateCurrentPeriodStart_BiWeekly(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{
			name:     "Week 1 (Jan 1, 2024 - Monday) should return Dec 31",
			date:     "2024-01-01", // Monday, Week 1
			expected: "2023-12-31", // Start of weeks 1-2 period (Sunday)
		},
		{
			name:     "Week 1 (Jan 7, 2024 - Sunday) should return Dec 31",
			date:     "2024-01-07", // Sunday, Week 1
			expected: "2023-12-31", // Start of weeks 1-2 period (Sunday)
		},
		{
			name:     "Week 2 (Jan 8, 2024 - Monday) should return Dec 31",
			date:     "2024-01-08", // Monday, Week 2
			expected: "2023-12-31", // Start of weeks 1-2 period (Sunday)
		},
		{
			name:     "Week 3 (Jan 15, 2024 - Monday) should return Jan 14",
			date:     "2024-01-15", // Monday, Week 3
			expected: "2024-01-14", // Start of weeks 3-4 period (Sunday)
		},
		{
			name:     "Week 4 (Jan 22, 2024 - Monday) should return Jan 14",
			date:     "2024-01-22", // Monday, Week 4
			expected: "2024-01-14", // Start of weeks 3-4 period (Sunday)
		},
		{
			name:     "Week 5 (Jan 29, 2024 - Monday) should return Jan 28",
			date:     "2024-01-29", // Monday, Week 5
			expected: "2024-01-28", // Start of weeks 5-6 period (Sunday)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			
			result := CalculateCurrentPeriodStart("bi-weekly", date)
			
			if !result.Equal(expected) {
				t.Errorf("CalculateCurrentPeriodStart() = %v, want %v", result.Format("2006-01-02"), expected.Format("2006-01-02"))
			}
		})
	}
}

func TestCalculateCurrentPeriodStart_Monthly(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{
			name:     "1st of January should return 1st of January",
			date:     "2024-01-01",
			expected: "2024-01-01",
		},
		{
			name:     "15th of January should return 1st of January",
			date:     "2024-01-15",
			expected: "2024-01-01",
		},
		{
			name:     "31st of January should return 1st of January",
			date:     "2024-01-31",
			expected: "2024-01-01",
		},
		{
			name:     "29th of February (leap year) should return 1st of February",
			date:     "2024-02-29",
			expected: "2024-02-01",
		},
		{
			name:     "31st of December should return 1st of December",
			date:     "2024-12-31",
			expected: "2024-12-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			
			result := CalculateCurrentPeriodStart("monthly", date)
			
			if !result.Equal(expected) {
				t.Errorf("CalculateCurrentPeriodStart() = %v, want %v", result.Format("2006-01-02"), expected.Format("2006-01-02"))
			}
		})
	}
}

func TestCalculateNextPeriodStart_Weekly(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{
			name:     "Monday should return next Sunday",
			date:     "2024-01-15", // Monday
			expected: "2024-01-21", // Next Sunday
		},
		{
			name:     "Friday should return next Sunday",
			date:     "2024-01-19", // Friday
			expected: "2024-01-21", // Next Sunday
		},
		{
			name:     "Sunday should return next Sunday",
			date:     "2024-01-21", // Sunday
			expected: "2024-01-28", // Next Sunday
		},
		{
			name:     "Year boundary - December Sunday",
			date:     "2023-12-31", // Sunday
			expected: "2024-01-07", // Next Sunday (New Year)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			
			result := CalculateNextPeriodStart("weekly", date)
			
			if !result.Equal(expected) {
				t.Errorf("CalculateNextPeriodStart() = %v, want %v", result.Format("2006-01-02"), expected.Format("2006-01-02"))
			}
		})
	}
}

func TestCalculateNextPeriodStart_BiWeekly(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{
			name:     "Week 1 (Jan 1) should return start of weeks 3-4 period",
			date:     "2024-01-01", // Monday, Week 1
			expected: "2024-01-14", // Sunday of Week 3
		},
		{
			name:     "Week 2 (Jan 8) should return start of weeks 3-4 period",
			date:     "2024-01-08", // Monday, Week 2
			expected: "2024-01-14", // Sunday of Week 3
		},
		{
			name:     "Week 3 (Jan 15) should return start of weeks 5-6 period",
			date:     "2024-01-15", // Monday, Week 3
			expected: "2024-01-28", // Sunday of Week 5
		},
		{
			name:     "Week 4 (Jan 22) should return start of weeks 5-6 period",
			date:     "2024-01-22", // Monday, Week 4
			expected: "2024-01-28", // Sunday of Week 5
		},
		{
			name:     "Week 5 (Jan 29) should return start of weeks 7-8 period",
			date:     "2024-01-29", // Monday, Week 5
			expected: "2024-02-11", // Sunday of Week 7
		},
		{
			name:     "Year boundary - Week 52 should return Week 1 of next year",
			date:     "2024-12-23", // Monday, Week 52 (in weeks 51-52 period)
			expected: "2024-12-29", // Sunday of Week 53 (which becomes Week 1 of next bi-weekly period)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			
			result := CalculateNextPeriodStart("bi-weekly", date)
			
			if !result.Equal(expected) {
				t.Errorf("CalculateNextPeriodStart() = %v, want %v", result.Format("2006-01-02"), expected.Format("2006-01-02"))
			}
		})
	}
}

func TestCalculateNextPeriodStart_Monthly(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{
			name:     "January should return February",
			date:     "2024-01-15",
			expected: "2024-02-01",
		},
		{
			name:     "February (leap year) should return March",
			date:     "2024-02-29",
			expected: "2024-03-01",
		},
		{
			name:     "February (non-leap year) should return March",
			date:     "2023-02-28",
			expected: "2023-03-01",
		},
		{
			name:     "December should return January (year boundary)",
			date:     "2024-12-31",
			expected: "2025-01-01",
		},
		{
			name:     "November should return December",
			date:     "2024-11-15",
			expected: "2024-12-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			
			result := CalculateNextPeriodStart("monthly", date)
			
			if !result.Equal(expected) {
				t.Errorf("CalculateNextPeriodStart() = %v, want %v", result.Format("2006-01-02"), expected.Format("2006-01-02"))
			}
		})
	}
}

func TestCalculatePeriodEnd_Weekly(t *testing.T) {
	tests := []struct {
		name        string
		periodStart string
		expected    string
	}{
		{
			name:        "Sunday start should end on Saturday",
			periodStart: "2024-01-14", // Sunday
			expected:    "2024-01-20", // Saturday
		},
		{
			name:        "Year boundary week",
			periodStart: "2023-12-31", // Sunday
			expected:    "2024-01-06", // Saturday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, _ := time.Parse("2006-01-02", tt.periodStart)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			// Set expected to end of day
			expected = time.Date(expected.Year(), expected.Month(), expected.Day(), 23, 59, 59, 0, expected.Location())
			
			result := CalculatePeriodEnd("weekly", start)
			
			if !result.Equal(expected) {
				t.Errorf("CalculatePeriodEnd() = %v, want %v", result, expected)
			}
		})
	}
}

func TestCalculatePeriodEnd_BiWeekly(t *testing.T) {
	tests := []struct {
		name        string
		periodStart string
		expected    string
	}{
		{
			name:        "1st of month should end on 14th",
			periodStart: "2024-01-01",
			expected:    "2024-01-14",
		},
		{
			name:        "15th of month should end on 28th",
			periodStart: "2024-01-15",
			expected:    "2024-01-28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, _ := time.Parse("2006-01-02", tt.periodStart)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			// Set expected to end of day
			expected = time.Date(expected.Year(), expected.Month(), expected.Day(), 23, 59, 59, 0, expected.Location())
			
			result := CalculatePeriodEnd("bi-weekly", start)
			
			if !result.Equal(expected) {
				t.Errorf("CalculatePeriodEnd() = %v, want %v", result, expected)
			}
		})
	}
}

func TestCalculatePeriodEnd_Monthly(t *testing.T) {
	tests := []struct {
		name        string
		periodStart string
		expected    string
	}{
		{
			name:        "January should end on January 31st",
			periodStart: "2024-01-01",
			expected:    "2024-01-31",
		},
		{
			name:        "February (leap year) should end on February 29th",
			periodStart: "2024-02-01",
			expected:    "2024-02-29",
		},
		{
			name:        "February (non-leap year) should end on February 28th",
			periodStart: "2023-02-01",
			expected:    "2023-02-28",
		},
		{
			name:        "April should end on April 30th",
			periodStart: "2024-04-01",
			expected:    "2024-04-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, _ := time.Parse("2006-01-02", tt.periodStart)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			// Set expected to end of day
			expected = time.Date(expected.Year(), expected.Month(), expected.Day(), 23, 59, 59, 0, expected.Location())
			
			result := CalculatePeriodEnd("monthly", start)
			
			if !result.Equal(expected) {
				t.Errorf("CalculatePeriodEnd() = %v, want %v", result, expected)
			}
		})
	}
}

func TestEdgeCases_LeapYear(t *testing.T) {
	tests := []struct {
		name     string
		function string
		freq     string
		date     string
		expected string
	}{
		{
			name:     "Leap year February 29th - current period start",
			function: "current",
			freq:     "monthly",
			date:     "2024-02-29",
			expected: "2024-02-01",
		},
		{
			name:     "Leap year February 29th - next period start",
			function: "next",
			freq:     "monthly",
			date:     "2024-02-29",
			expected: "2024-03-01",
		},
		{
			name:     "Non-leap year February 28th - next period start",
			function: "next",
			freq:     "monthly",
			date:     "2023-02-28",
			expected: "2023-03-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			
			var result time.Time
			if tt.function == "current" {
				result = CalculateCurrentPeriodStart(tt.freq, date)
			} else {
				result = CalculateNextPeriodStart(tt.freq, date)
			}
			
			if !result.Equal(expected) {
				t.Errorf("Function %s with %s frequency = %v, want %v", tt.function, tt.freq, result.Format("2006-01-02"), expected.Format("2006-01-02"))
			}
		})
	}
}

func TestEdgeCases_YearBoundary(t *testing.T) {
	tests := []struct {
		name     string
		function string
		freq     string
		date     string
		expected string
	}{
		{
			name:     "December 31st - next weekly period",
			function: "next",
			freq:     "weekly",
			date:     "2023-12-31", // Sunday
			expected: "2024-01-07", // Next Sunday
		},
		{
			name:     "December 31st - next monthly period",
			function: "next",
			freq:     "monthly",
			date:     "2023-12-31",
			expected: "2024-01-01",
		},
		{
			name:     "December 31st - next bi-weekly period",
			function: "next",
			freq:     "bi-weekly",
			date:     "2023-12-31",
			expected: "2023-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			expected, _ := time.Parse("2006-01-02", tt.expected)
			
			var result time.Time
			if tt.function == "current" {
				result = CalculateCurrentPeriodStart(tt.freq, date)
			} else {
				result = CalculateNextPeriodStart(tt.freq, date)
			}
			
			if !result.Equal(expected) {
				t.Errorf("Function %s with %s frequency = %v, want %v", tt.function, tt.freq, result.Format("2006-01-02"), expected.Format("2006-01-02"))
			}
		})
	}
}

func TestInvalidFrequency(t *testing.T) {
	date := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	
	// Test that invalid frequency defaults to monthly
	currentResult := CalculateCurrentPeriodStart("invalid", date)
	expectedCurrent := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	if !currentResult.Equal(expectedCurrent) {
		t.Errorf("CalculateCurrentPeriodStart with invalid frequency = %v, want %v", currentResult, expectedCurrent)
	}
	
	nextResult := CalculateNextPeriodStart("invalid", date)
	expectedNext := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	
	if !nextResult.Equal(expectedNext) {
		t.Errorf("CalculateNextPeriodStart with invalid frequency = %v, want %v", nextResult, expectedNext)
	}
}

func TestTimeZoneHandling(t *testing.T) {
	// Test with different time zones to ensure calculations work correctly
	est, _ := time.LoadLocation("America/New_York")
	pst, _ := time.LoadLocation("America/Los_Angeles")
	
	tests := []struct {
		name     string
		date     time.Time
		expected string // Use string for easier comparison
	}{
		{
			name:     "EST timezone - weekly calculation",
			date:     time.Date(2024, 1, 17, 15, 30, 0, 0, est), // Wednesday in EST
			expected: "2024-01-14", // Sunday
		},
		{
			name:     "PST timezone - weekly calculation",
			date:     time.Date(2024, 1, 17, 12, 30, 0, 0, pst), // Wednesday in PST
			expected: "2024-01-14", // Sunday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCurrentPeriodStart("weekly", tt.date)
			
			// Compare just the date part, not the exact time/timezone
			if result.Format("2006-01-02") != tt.expected {
				t.Errorf("CalculateCurrentPeriodStart() = %v, want %v", result.Format("2006-01-02"), tt.expected)
			}
			
			// Ensure the result maintains the same timezone as input
			if result.Location() != tt.date.Location() {
				t.Errorf("CalculateCurrentPeriodStart() timezone = %v, want %v", result.Location(), tt.date.Location())
			}
		})
	}
}