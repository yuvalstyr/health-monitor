package timeutil

import (
	"time"
)

// CalculateCurrentPeriodStart calculates the start date of the current period
// based on the frequency and current time
func CalculateCurrentPeriodStart(frequency string, currentTime time.Time) time.Time {
	switch frequency {
	case "weekly":
		// Sunday of current week
		weekday := int(currentTime.Weekday()) // Sunday = 0, Monday = 1, etc.
		sunday := currentTime.AddDate(0, 0, -weekday)
		// Set to start of day in the same timezone
		return time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 0, 0, 0, 0, sunday.Location())

	case "bi-weekly":
		// Calculate based on ISO week numbers: weeks 1-2, 3-4, 5-6, etc.
		year, week := currentTime.ISOWeek()
		
		// Determine which bi-weekly period this week belongs to
		// Weeks 1-2 = period 1, weeks 3-4 = period 2, etc.
		var periodStartWeek int
		if week%2 == 1 {
			// Odd week number (1, 3, 5, ...) - start of bi-weekly period
			periodStartWeek = week
		} else {
			// Even week number (2, 4, 6, ...) - second week of bi-weekly period
			periodStartWeek = week - 1
		}
		
		// Calculate the Sunday of the period start week
		jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, currentTime.Location())
		jan1Weekday := int(jan1.Weekday())
		
		// Find the Monday of week 1 (ISO week standard)
		week1Monday := jan1.AddDate(0, 0, -(jan1Weekday-1))
		if jan1Weekday == 0 {
			week1Monday = jan1.AddDate(0, 0, 1) // If Jan 1 is Sunday, Monday is next day
		}
		if jan1Weekday > 4 { // If Jan 1 is Fri, Sat, or Sun, week 1 starts next Monday
			week1Monday = week1Monday.AddDate(0, 0, 7)
		}
		
		// Calculate the Monday of our target week, then go back to Sunday
		periodStartMonday := week1Monday.AddDate(0, 0, (periodStartWeek-1)*7)
		periodStartSunday := periodStartMonday.AddDate(0, 0, -1)
		return periodStartSunday

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
		// Sunday of next week
		currentStart := CalculateCurrentPeriodStart(frequency, currentTime)
		return currentStart.AddDate(0, 0, 7)

	case "bi-weekly":
		// Next bi-weekly period (2 weeks after current period start)
		currentStart := CalculateCurrentPeriodStart(frequency, currentTime)
		return currentStart.AddDate(0, 0, 14)

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
		// Saturday of the same week (6 days after Sunday)
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