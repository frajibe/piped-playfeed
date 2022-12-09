package model

import "strings"

var SyncDurationUnitMonth = "month"
var SyncDurationUnitDay = "day"

type Duration struct {
	Unit  string `validate:"oneof=day month"`
	Value int    `validate:"number,min=1,max=12"`
}

func (duration *Duration) SetDefaults() {
	if strings.TrimSpace(duration.Unit) == "" {
		duration.Unit = SyncDurationUnitMonth
	}
	if duration.Value == 0 {
		duration.Value = 1
	}
}
