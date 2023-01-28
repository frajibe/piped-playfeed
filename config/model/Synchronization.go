package model

import (
	"strings"
	"time"
)

var defaultPlaylistPrefix = "PF - "
var SyncDurationType = "duration"
var SyncDateType = "date"

var PlaylistWeeklyStrategy = "week"
var PlaylistMonthlyStrategy = "month"

type Synchronization struct {
	Strategy       string `validate:"oneof=week month"`
	PlaylistPrefix string
	Type           string   `validate:"oneof=date duration"`
	Date           string   `validate:"datetime=2006-01-02,dateinpast"`
	Duration       Duration `validate:"required"`
}

func (synchronization *Synchronization) SetDefaults() {
	if strings.TrimSpace(synchronization.PlaylistPrefix) == "" {
		synchronization.PlaylistPrefix = defaultPlaylistPrefix
	}
	if strings.TrimSpace(synchronization.Strategy) == "" {
		synchronization.Strategy = PlaylistMonthlyStrategy
	}
	if strings.TrimSpace(synchronization.Type) == "" {
		synchronization.Type = SyncDurationType
	}
	if strings.TrimSpace(synchronization.Date) == "" {
		synchronization.Date = time.Now().Local().AddDate(0, -1, 0).Format("2006-01-02")
	}
	synchronization.Duration.SetDefaults()
}
