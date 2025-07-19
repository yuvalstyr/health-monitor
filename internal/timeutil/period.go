package timeutil

import (
	"time"
)

// CalculateCurrentPeriodStart calculates the start date of the current period
// based on the frequency and current time
func CalculateCurrentPeriodStart(frequency string, currentTime time.Time) time.Time {
	switch frequency {
	case "weekly":
		// Monday of current week
		weekday := int(currentTime.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		monday := currentTime.AddDate(0, 0, -(weekday-1))
		// Set to start of day in the same timezone
		return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())

	case "bi-weekly":
		// First or 15th of current month based on current date
		if currentTime.Day() <= 14 {
			return time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
		} else {
			return time.Date(currentTime.Year(), currentTime.Month(), 15, 0, 0, 0, 0, currentTime.Location())
		}

	case "monthly":
		// First day of current month
		return time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())

	default:
		// Default to monthly for unknown frequencies
		return time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
	}
}

// CalculateNextPeriodStart calculates the start date of the next period
// based on the frequency and current time
func CalculateNextPeriodStart(frequency string, currentTime time.Time) time.Time {
	switch frequency {
	case "weekly":
		// Monday of next week
		currentStart := CalculateCurrentPeriodStart(frequency, currentTime)
		return currentStart.AddDate(0, 0, 7)

	case "bi-weekly":
		// Next bi-weekly period
		if currentTime.Day() <= 14 {
			// Currently in first half, next is 15th
			return time.Date(currentTime.Year(), currentTime.Month(), 15, 0, 0, 0, 0, currentTime.Location())
		} else {
			// Currently in second half, next is 1st of next month
			return time.Date(currentTime.Year(), currentTime.Month()+1, 1, 0, 0, 0, 0, currentTime.Location())
		}

	case "monthly":
		// First day of next month
		return time.Date(currentTime.Year(), currentTime.Month()+1, 1, 0, 0, 0, 0, currentTime.Location())

	default:
		// Default to monthly for unknown frequencies
		return time.Date(currentTime.Year(), currentTime.Month()+1, 1, 0, 0, 0, 0, currentTime.Location())
	}
}

// CalculatePeriodEnd calculates the end date of a period given its start date and frequency
func CalculatePeriodEnd(frequency string, periodStart time.Time) time.Time {
	switch frequency {
	case "weekly":
		// Sunday of the same week (6 days after Monday)
		return periodStart.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	case "bi-weekly":
		// 13 days after start (14 days total)
		return periodStart.AddDate(0, 0, 13).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	case "monthly":
		// Last day of the same month
		nextMonth := periodStart.AddDate(0, 1, 0)
		lastDay := nextMonth.AddDate(0, 0, -1)
		return time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), 23, 59, 59, 0, periodStart.Location())

	default:
		// Default to monthly
		nextMonth := periodStart.AddDate(0, 1, 0)
		lastDay := nextMonth.AddDate(0, 0, -1)
		return time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), 23, 59, 59, 0, periodStart.Location())
	}
}