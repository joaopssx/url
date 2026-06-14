package model

import "time"

type DayStat struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type WeekStat struct {
	Week  string `json:"week"`
	Count int    `json:"count"`
}

type MonthlyStat struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type StatsResult struct {
	ShortCode     string        `json:"short_code"`
	OriginalURL   string        `json:"original_url"`
	TotalAccesses int           `json:"total_accesses"`
	CreatedAt     time.Time     `json:"created_at"`
	ExpiresAt     *time.Time    `json:"expires_at"`
	Daily         []DayStat     `json:"daily"`
	Weekly        []WeekStat    `json:"weekly"`
	Monthly       []MonthlyStat `json:"monthly"`
}
