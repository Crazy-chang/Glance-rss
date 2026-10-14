package fetcher

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"glance-wx/backend/model"
)

// 真实抓取集成测试，默认跳过；需要联网时执行：
//
//	GLANCE_LIVE=1 go test ./fetcher/ -run TestLiveFetch -v
func TestLiveFetch(t *testing.T) {
	if os.Getenv("GLANCE_LIVE") != "1" {
		t.Skip("跳过真实抓取测试，设置 GLANCE_LIVE=1 后执行")
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	if err := model.AutoMigrate(db); err != nil {
		t.Fatalf("建表失败: %v", err)
	}

	user := model.User{OpenID: "live-test-openid"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("建用户失败: %v", err)
	}
	setting := model.DefaultSetting(user.ID)
	setting.MaxItems = 5 // 故意调小，验证裁剪
	if err := db.Create(&setting).Error; err != nil {
		t.Fatalf("建设置失败: %v", err)
	}

	src := model.Source{
		UserID:  user.ID,
		Title:   "https://www.ruanyifeng.com/blog/atom.xml", // 故意先放 URL，验证元数据回填
		FeedURL: "https://www.ruanyifeng.com/blog/atom.xml",
	}
	if err := db.Create(&src).Error; err != nil {
		t.Fatalf("建订阅源失败: %v", err)
	}

	f := New(db, 15*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := f.FetchSource(ctx, src.ID); err != nil {
		t.Fatalf("抓取失败: %v", err)
	}

	db.First(&src, src.ID)
	if src.Title == "" || src.Title == src.FeedURL {
		t.Errorf("标题未被 feed 元数据回填: %q", src.Title)
	}
	if src.SiteURL == "" {
		t.Error("站点地址未回填")
	}
	if src.LastFetchAt == nil {
		t.Error("last_fetch_at 未更新")
	}
	if src.LastError != "" {
		t.Errorf("last_error 应为空，实际: %s", src.LastError)
	}
	t.Logf("源信息: title=%q site=%q favicon=%q", src.Title, src.SiteURL, src.FaviconPath)

	var articles []model.Article
	db.Where("source_id = ?", src.ID).Order("published_at DESC").Find(&articles)
	if len(articles) == 0 {
		t.Fatal("未抓到任何文章")
	}
	if len(articles) > 5 {
		t.Errorf("裁剪失败，文章数 %d 超过最大条数 5", len(articles))
	}
	first := articles[0]
	if first.Title == "" {
		t.Error("文章标题为空")
	}
	if first.Summary == "" {
		t.Error("摘要为空")
	}
	if first.Content == "" {
		t.Error("正文为空")
	}
	if !first.PublishedAt.IsZero() && time.Since(first.PublishedAt) > 365*24*time.Hour {
		t.Errorf("发布时间异常: %v", first.PublishedAt)
	}
	t.Logf("抓到 %d 篇，首篇: %s | 摘要: %.40s", len(articles), first.Title, first.Summary)

	// 再抓一次，验证去重
	if err := f.FetchSource(ctx, src.ID); err != nil {
		t.Fatalf("第二次抓取失败: %v", err)
	}
	var count int64
	db.Model(&model.Article{}).Where("source_id = ?", src.ID).Count(&count)
	if int(count) != len(articles) {
		t.Errorf("去重失败，第一次 %d 篇，第二次 %d 篇", len(articles), count)
	}
}
