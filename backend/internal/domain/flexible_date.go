package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type FlexibleDate struct {
	time.Time
}

// NewFlexibleDate returns a FlexibleDate wrapping the given time.
func NewFlexibleDate(t time.Time) FlexibleDate {
	return FlexibleDate{Time: t}
}

// UnmarshalJSON handles both "2006-01-02" and RFC3339 formats
func (fd *FlexibleDate) UnmarshalJSON(b []byte) error {
	str := string(b)
	str = str[1 : len(str)-1] // remove quotes

	// Try parsing as "2006-01-02"
	if t, err := time.Parse("2006-01-02", str); err == nil {
		fd.Time = t
		return nil
	}

	// Try RFC3339 fallback
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		fd.Time = t
		return nil
	}

	return fmt.Errorf("invalid date format: %s", str)
}

// MarshalJSON ensures we always serialize in "2006-01-02" format
func (fd FlexibleDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(fd.Format("2006-01-02"))
}

// IsZero allows comparison and validation
func (fd FlexibleDate) IsZero() bool {
	return fd.Time.IsZero()
}
