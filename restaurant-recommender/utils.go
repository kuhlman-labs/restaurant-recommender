package restaurantrecommender

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseQuery extracts filtering criteria from a free‑form query using the provided styles.
func parseQuery(query string, styles []string) QueryCriteria {
	criteria := QueryCriteria{}
	lowerQuery := strings.ToLower(query)

	// Dynamically detect the restaurant style.
	for _, style := range styles {
		if strings.Contains(lowerQuery, strings.ToLower(style)) {
			criteria.Style = style
			break
		}
	}

	// Check for "vegetarian".
	if strings.Contains(lowerQuery, "vegetarian") {
		val := true
		criteria.Vegetarian = &val
	}

	// Check for delivery keywords.
	if strings.Contains(lowerQuery, "deliver") {
		val := true
		criteria.Delivers = &val
	}

	// Check for "open now".
	if strings.Contains(lowerQuery, "open now") {
		criteria.OpenNow = true
	}

	// Check for a specific time e.g. "open at 6pm" or "open at 6:30pm".
	re := regexp.MustCompile(`open at (\d{1,2})(?::(\d{2}))?\s*(am|pm)`)
	if matches := re.FindStringSubmatch(lowerQuery); matches != nil {
		hour, _ := strconv.Atoi(matches[1])
		minute := 0
		if matches[2] != "" {
			minute, _ = strconv.Atoi(matches[2])
		}
		ampm := matches[3]
		if ampm == "pm" && hour != 12 {
			hour += 12
		} else if ampm == "am" && hour == 12 {
			hour = 0
		}
		now := time.Now()
		openAt := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		criteria.OpenAt = &openAt
	}

	return criteria
}

// parseTime converts a string (e.g., "09:00") to a time.Time object.
func parseTime(tStr string) (time.Time, error) {
	return time.Parse("15:04", tStr)
}

// isOpen checks if a restaurant is open at the specified time.
func isOpen(r Restaurant, currentTime time.Time) bool {
	open, err := parseTime(r.OpenHour)
	if err != nil {
		return false
	}
	close, err := parseTime(r.CloseHour)
	if err != nil {
		return false
	}

	now := time.Date(0, 1, 1, currentTime.Hour(), currentTime.Minute(), 0, 0, time.UTC)
	openTime := time.Date(0, 1, 1, open.Hour(), open.Minute(), 0, 0, time.UTC)
	closeTime := time.Date(0, 1, 1, close.Hour(), close.Minute(), 0, 0, time.UTC)

	// Adjust for restaurants that close past midnight.
	if closeTime.Before(openTime) {
		if now.Before(openTime) {
			now = now.Add(24 * time.Hour)
		}
		closeTime = closeTime.Add(24 * time.Hour)
	}
	return now.After(openTime) && now.Before(closeTime)
}

// restaurantMatchesCriteria returns true if a restaurant meets the query criteria.
func restaurantMatchesCriteria(r Restaurant, criteria QueryCriteria, now time.Time) bool {
	if criteria.Style != "" && strings.ToLower(r.Style) != strings.ToLower(criteria.Style) {
		return false
	}
	if criteria.Vegetarian != nil && r.Vegetarian != *criteria.Vegetarian {
		return false
	}
	if criteria.Delivers != nil && r.Deliveries != *criteria.Delivers {
		return false
	}
	var checkTime time.Time
	if criteria.OpenAt != nil {
		checkTime = *criteria.OpenAt
	} else if criteria.OpenNow {
		checkTime = now
	}
	if !checkTime.IsZero() && !isOpen(r, checkTime) {
		return false
	}
	return true
}
