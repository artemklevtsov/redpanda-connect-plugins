package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// ParseDate parses a date string and returns a formatted date string.
// It supports "today", "yesterday", "tomorrow", "Ndaysago" and "YYYY-MM-DD" formats.
func ParseDate(s string) (string, error) {
	s = strings.ToLower(s)
	switch s {
	case "today":
		return time.Now().
			Format(dateLayout), nil
	case "yesterday":
		return time.Now().
			AddDate(0, 0, -1).
			Format(dateLayout), nil
	case "tomorrow":
		return time.Now().
			AddDate(0, 0, 1).
			Format(dateLayout), nil
	default:
		if strings.HasSuffix(s, "daysago") {
			num, err := strconv.Atoi(strings.TrimSuffix(s, "daysago"))
			if err != nil {
				return "", fmt.Errorf("cannot parse %q: invalid daysago format (NdaysAgo)", s)
			}

			return time.Now().
				AddDate(0, 0, -num).
				Format(dateLayout), nil
		}

		d, err := time.Parse(dateLayout, s)
		if err != nil {
			return "", fmt.Errorf("cannot parse %q: invalid date format (YYYY-MM-DD)", s)
		}

		return d.Format(dateLayout), nil
	}
}
