package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"glance-wx/backend/fetcher"
	"glance-wx/backend/model"
)

// sourceDueRow 到期检查查询行
type sourceDueRow struct {
	ID          int64
	LastFetchAt *time.Time
}

// Scheduler 定时调度器：每分钟检查一次，到期的源交给 fetcher 并发抓取
type Scheduler struct {
	db              *gorm.DB
	fetcher         *fetcher.Fetcher
	concurrency     int
	refreshInterval string
	cron            *cron.Cron
}

func New(db *gorm.DB, f *fetcher.Fetcher, concurrency int, refreshInterval string) *Scheduler {
	return &Scheduler{
		db:              db,
		fetcher:         f,
		concurrency:     concurrency,
		refreshInterval: refreshInterval,
	}
}

// Start 启动调度：立即跑一轮，之后每分钟检查一轮
func (s *Scheduler) Start(ctx context.Context) {
	s.cron = cron.New()
	_, _ = s.cron.AddFunc("@every 1m", func() {
		s.runOnce(ctx)
	})
	s.cron.Start()

	// 启动后先跑一轮，让新添加/从未抓取的源尽快入库，不用干等 1 分钟
	go s.runOnce(ctx)
	slog.Info("定时抓取调度器已启动", "check_interval", "1m", "refresh_interval", s.refreshInterval, "concurrency", s.concurrency)
}

// Stop 停止调度
func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

// runOnce 查出全部源，按全局间隔筛出到期源并发抓取
func (s *Scheduler) runOnce(ctx context.Context) {
	var rows []sourceDueRow
	err := s.db.Table("sources").
		Select("id AS id, last_fetch_at AS last_fetch_at").
		Scan(&rows).Error
	if err != nil {
		slog.Warn("调度器查询订阅源失败", "err", err)
		return
	}

	now := time.Now()
	var dueIDs []int64
	for _, r := range rows {
		if IsDue(r.LastFetchAt, s.refreshInterval, now) {
			dueIDs = append(dueIDs, r.ID)
		}
	}
	if len(dueIDs) == 0 {
		return
	}

	slog.Info("调度器本轮到期订阅源", "count", len(dueIDs))

	sem := make(chan struct{}, s.concurrency)
	var wg sync.WaitGroup
	for _, id := range dueIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(sourceID int64) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := s.fetcher.FetchSource(ctx, sourceID); err != nil {
				slog.Warn("定时抓取失败", "source_id", sourceID, "err", err)
			}
		}(id)
	}
	wg.Wait()
}

// RefreshUserSources 手动触发：刷新某用户的全部订阅源（异步，不阻塞调用方）
func (s *Scheduler) RefreshUserSources(ctx context.Context, userID int64) {
	var ids []int64
	if err := s.db.Model(&model.Source{}).
		Where("user_id = ?", userID).
		Pluck("id", &ids).Error; err != nil {
		slog.Warn("查询用户订阅源失败", "user_id", userID, "err", err)
		return
	}
	if len(ids) == 0 {
		return
	}

	sem := make(chan struct{}, s.concurrency)
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		sem <- struct{}{}
		go func(sourceID int64) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := s.fetcher.FetchSource(ctx, sourceID); err != nil {
				slog.Warn("手动刷新抓取失败", "source_id", sourceID, "err", err)
			}
		}(id)
	}
	go func() {
		wg.Wait()
		slog.Info("用户手动刷新完成", "user_id", userID, "sources", len(ids))
	}()
}
