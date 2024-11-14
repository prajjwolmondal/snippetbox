package main

import (
	"testing"
	"time"

	"snippetbox.prajjmon.net/internal/assert"
)

func TestHumanDate(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2024, 11, 12, 10, 15, 0, 0, time.UTC),
			want: "12 Nov 2024 at 10:15",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "EST",
			tm:   time.Date(2024, 11, 12, 10, 15, 0, 0, time.FixedZone("EST", -5*60*60)),
			want: "12 Nov 2024 at 15:15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)

			assert.Equal(t, hd, tt.want)
		})
	}
}
