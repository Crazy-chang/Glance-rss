// Package scheduler 每分钟检查到期订阅源并触发抓取
package scheduler

import (
	"time"

	"glance-wx/backend/model"
)

// intervalMinutes 把设置里的间隔编码换算为分钟；
// 返回 ok=false 表示"手动"或无法识别，不参与定时抓取
func intervalMinutes(interval string) (minutes int, ok bool) {
	switch interval {
	case model.RefreshInterval30m:
		return 30, true
	case model.RefreshInterval1h:
		return 60, true
	case model.RefreshInterval2h:
		return 120, true
	default: // manual 及未知值
		return 0, false
	}
}

// IsDue 判断某订阅源在 now 时刻是否到期应抓取：
//   - 从未抓取过（lastFetchAt 为空）：到期
//   - 手动/未知间隔：永不到期
//   - 否则：lastFetchAt + 间隔 <= now
func IsDue(lastFetchAt *time.Time, interval string, now time.Time) bool {
	minutes, ok := intervalMinutes(interval)
	if !ok {
		return false
	}
	if lastFetchAt == nil {
		return true
	}
	return !lastFetchAt.Add(time.Duration(minutes) * time.Minute).After(now)
}
