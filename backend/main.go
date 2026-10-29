// Glance RSS 阅读小程序后端入口
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/config"
	"glance-wx/backend/fetcher"
	"glance-wx/backend/handler"
	"glance-wx/backend/model"
	"glance-wx/backend/scheduler"
	"glance-wx/backend/service"
)

func main() {
	cfg := config.MustLoad()

	// 数据库：打开 → 建表 → 写入推荐源种子
	db, err := model.Open(cfg.Database.Path)
	if err != nil {
		slog.Error("数据库初始化失败", "err", err)
		os.Exit(1)
	}
	if err := model.AutoMigrate(db); err != nil {
		slog.Error("建表失败", "err", err)
		os.Exit(1)
	}
	if err := model.SeedRecommendSources(db); err != nil {
		slog.Error("推荐源种子写入失败", "err", err)
		os.Exit(1)
	}
	// 客户端不再开放条数修改，存量用户统一归一化为默认值
	if err := db.Model(&model.Setting{}).Where("max_items <> ?", model.DefaultMaxItems).
		Update("max_items", model.DefaultMaxItems).Error; err != nil {
		slog.Error("统一单源最大条数失败", "err", err)
		os.Exit(1)
	}
	slog.Info("数据库就绪", "path", cfg.Database.Path)

	// 抓取器与定时调度器
	f := fetcher.New(db, time.Duration(cfg.Fetch.TimeoutSeconds)*time.Second)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sched := scheduler.New(db, f, cfg.Fetch.Concurrency, cfg.Server.RefreshInterval)
	sched.Start(ctx)
	defer sched.Stop()

	// 业务 service 装配
	authSvc := service.NewAuthService(db, cfg)
	sourceSvc := service.NewSourceService(db, f, cfg.Server.BaseURL)
	h := handler.New(handler.Deps{
		Cfg:      cfg,
		Auth:     authSvc,
		Article:  service.NewArticleService(db, cfg.Server.BaseURL, cfg.JWT.Secret),
		Source:   sourceSvc,
		Discover: service.NewDiscoverService(db, sourceSvc),
		Setting:  service.NewSettingService(db),
		Proxy:    service.NewProxyService(),
		Sched:    sched,
	})

	r := gin.Default()
	h.Register(r.Group("/api"))

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		slog.Info("Glance 后端启动", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("服务启动失败", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("收到退出信号，正在关闭服务...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
