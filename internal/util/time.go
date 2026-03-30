package util

import "time"

func UTCNowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
