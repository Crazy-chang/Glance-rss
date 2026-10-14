package scheduler

import (
	"testing"
	"time"

	"glance-wx/backend/model"
)

func ptrTime(t time.Time) *time.Time { return &t }

func TestIsDue(t *testing.T) {
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.Local)

	cases := []struct {
		name     string
		last     *time.Time
		interval string
		want     bool
	}{
		{"从未抓取-30分钟", nil, model.RefreshInterval30m, true},
		{"从未抓取-1小时", nil, model.RefreshInterval1h, true},
		{"手动间隔-从未抓取", nil, model.RefreshIntervalManual, false},
		{"手动间隔-曾经抓取", ptrTime(now.Add(-3 * time.Hour)), model.RefreshIntervalManual, false},
		{"20分钟前-间隔30分钟-未到期", ptrTime(now.Add(-20 * time.Minute)), model.RefreshInterval30m, false},
		{"30分钟前-间隔30分钟-到期", ptrTime(now.Add(-30 * time.Minute)), model.RefreshInterval30m, true},
		{"40分钟前-间隔30分钟-到期", ptrTime(now.Add(-40 * time.Minute)), model.RefreshInterval30m, true},
		{"61分钟前-间隔1小时-到期", ptrTime(now.Add(-61 * time.Minute)), model.RefreshInterval1h, true},
		{"59分钟前-间隔1小时-未到期", ptrTime(now.Add(-59 * time.Minute)), model.RefreshInterval1h, false},
		{"3小时前-间隔2小时-到期", ptrTime(now.Add(-3 * time.Hour)), model.RefreshInterval2h, true},
		{"未来时间-未到期", ptrTime(now.Add(10 * time.Minute)), model.RefreshInterval30m, false},
		{"未知间隔-不到期", ptrTime(now.Add(-24 * time.Hour)), "weird", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsDue(c.last, c.interval, now); got != c.want {
				t.Errorf("IsDue() = %v, want %v", got, c.want)
			}
		})
	}
}
