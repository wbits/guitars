package persistence

import "time"

const timeRFC3339Nano = time.RFC3339Nano

func parseTimeRFC3339(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, raw)
}
