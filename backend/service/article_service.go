package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"gorm.io/gorm"

	"glance-wx/backend/model"
)

// 文章筛选类型
const (
	FilterAll       = "all"
	FilterUnread    = "unread"
	FilterRead      = "read"
	FilterToday     = "today"
	FilterFavorited = "favorited"
)

// chinaZone 文章"今日"按用户所在时区（UTC+8）计算
var chinaZone = time.FixedZone("CST", 8*60*60)

// ArticleService 文章查询与状态变更
// secret 用于分享链接的 HMAC 签名（复用 JWT 密钥）
type ArticleService struct {
	db      *gorm.DB
	baseURL string
	secret  string
}

func NewArticleService(db *gorm.DB, baseURL, secret string) *ArticleService {
	return &ArticleService{db: db, baseURL: strings.TrimRight(baseURL, "/"), secret: secret}
}

// ArticleListItem 列表项（不含正文）
type ArticleListItem struct {
	ID          int64  `json:"id"`
	SourceID    int64  `json:"source_id"`
	SourceName  string `json:"source_name"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Author      string `json:"author"`
	Link        string `json:"link"`
	PublishedAt int64  `json:"published_at"` // 秒级时间戳
	IsRead      bool   `json:"is_read"`
	IsFavorited bool   `json:"is_favorited"`
}

// ArticleDetail 详情（含正文，正文图片地址已重写为 img-proxy）
type ArticleDetail struct {
	ArticleListItem
	Content string `json:"content"`
	// ShareSig 分享签名：前端拼进分享链接，验签通过方可免订阅阅读
	ShareSig string `json:"share_sig"`
}

// ArticleListParams 列表查询参数
type ArticleListParams struct {
	UserID   int64
	Filter   string
	Search   string
	SourceID int64
	Page     int
	Size     int
}

// articleRow DB 查询中间结构
type articleRow struct {
	ID          int64
	SourceID    int64
	SourceName  string
	Title       string
	Summary     string
	Author      string
	Link        string
	PublishedAt time.Time
	IsRead      bool
	IsFavorited bool
	Content     string
}

// List 分页查询当前用户的文章
func (s *ArticleService) List(p ArticleListParams) (PageResult, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 50
	}
	if p.Size > 50 {
		p.Size = 50
	}

	q := s.buildListQuery(p.UserID, p.Filter, p.Search, p.SourceID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return PageResult{}, err
	}

	var setting model.Setting
	order := "a.published_at DESC, a.id DESC"
	if err := s.db.Where("user_id = ?", p.UserID).First(&setting).Error; err == nil {
		if setting.SortOrder == model.SortOrderAsc {
			order = "a.published_at ASC, a.id ASC"
		}
	}

	var rows []articleRow
	err := q.Select(`
			a.id, a.source_id, src.title AS source_name, a.title, a.summary,
			a.author, a.link, a.published_at, a.content,
			COALESCE(st.is_read, 0) AS is_read,
			COALESCE(st.is_favorited, 0) AS is_favorited`).
		Order(order).
		Offset((p.Page - 1) * p.Size).
		Limit(p.Size).
		Scan(&rows).Error
	if err != nil {
		return PageResult{}, err
	}

	list := make([]ArticleListItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, toListItem(r))
	}
	return NewPageResult(list, total, p.Page, p.Size), nil
}

// Detail 查询单篇文章详情（含正文）
func (s *ArticleService) Detail(userID, articleID int64) (*ArticleDetail, error) {
	var r articleRow
	err := s.baseScopedQuery(userID).
		Select(`
			a.id, a.source_id, src.title AS source_name, a.title, a.summary,
			a.author, a.link, a.published_at, a.content,
			COALESCE(st.is_read, 0) AS is_read,
			COALESCE(st.is_favorited, 0) AS is_favorited`).
		Where("a.id = ?", articleID).
		Take(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleMissing
	}
	if err != nil {
		return nil, err
	}

	item := toListItem(r)
	return &ArticleDetail{
		ArticleListItem: item,
		Content:         RewriteImageProxy(r.Content, s.baseURL+"/api/img-proxy?url="),
		ShareSig:        s.signArticle(articleID),
	}, nil
}

// signArticle 用密钥对文章 id 做 HMAC-SHA256 签名，取前 16 位 hex。
// 没有密钥遍历 id 也算不出合法签名，防止分享接口被当作全库公开读取。
func (s *ArticleService) signArticle(articleID int64) string {
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write([]byte("article:" + strconv.FormatInt(articleID, 10)))
	return hex.EncodeToString(mac.Sum(nil))[:16]
}

// SharedDetail 分享链路详情：验签通过后跳过用户订阅隔离，按文章 id 全局查。
// 对方没有该文章的状态记录，已读/收藏固定返回 false。
func (s *ArticleService) SharedDetail(articleID int64, sig string) (*ArticleDetail, error) {
	if !hmac.Equal([]byte(sig), []byte(s.signArticle(articleID))) {
		return nil, ErrArticleMissing
	}
	var r articleRow
	err := s.db.Table("articles a").
		Select(`
			a.id, a.source_id, src.title AS source_name, a.title, a.summary,
			a.author, a.link, a.published_at, a.content,
			0 AS is_read, 0 AS is_favorited`).
		Joins("JOIN sources src ON src.id = a.source_id").
		Where("a.id = ?", articleID).
		Take(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleMissing
	}
	if err != nil {
		return nil, err
	}
	item := toListItem(r)
	return &ArticleDetail{
		ArticleListItem: item,
		Content:         RewriteImageProxy(r.Content, s.baseURL+"/api/img-proxy?url="),
		ShareSig:        s.signArticle(articleID),
	}, nil
}

// SetRead 标记已读/未读
func (s *ArticleService) SetRead(userID, articleID int64, read bool) error {
	if err := s.checkOwnership(userID, articleID); err != nil {
		return err
	}
	return s.upsertState(userID, articleID, func(st *model.UserArticleState) {
		st.IsRead = read
		if read {
			now := time.Now().UTC()
			st.ReadAt = &now
		} else {
			st.ReadAt = nil
		}
	})
}

// SetFavorite 收藏/取消收藏
func (s *ArticleService) SetFavorite(userID, articleID int64, favorited bool) error {
	if err := s.checkOwnership(userID, articleID); err != nil {
		return err
	}
	return s.upsertState(userID, articleID, func(st *model.UserArticleState) {
		st.IsFavorited = favorited
	})
}

// ClearRead 清理当前用户已读且未收藏的文章，返回删除的文章数量。
// 收藏的文章（即使已读）会被保留，避免误删用户主动标记的内容。
func (s *ArticleService) ClearRead(userID int64) (int64, error) {
	var deleted int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 找出该用户已读且未收藏的文章 ID
		var ids []int64
		if err := tx.Table("user_article_states st").
			Joins("JOIN articles a ON a.id = st.article_id").
			Joins("JOIN sources src ON src.id = a.source_id").
			Where("st.user_id = ? AND src.user_id = ? AND st.is_read = 1 AND st.is_favorited = 0", userID, userID).
			Pluck("a.id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		// 删除文章
		res := tx.Where("id IN ?", ids).Delete(&model.Article{})
		if res.Error != nil {
			return res.Error
		}
		deleted = res.RowsAffected
		// 删除对应的用户状态记录
		tx.Where("article_id IN ? AND user_id = ?", ids, userID).Delete(&model.UserArticleState{})
		return nil
	})
	return deleted, err
}

// checkOwnership 校验文章属于该用户的订阅源
func (s *ArticleService) checkOwnership(userID, articleID int64) error {
	var count int64
	err := s.db.Table("articles a").
		Joins("JOIN sources src ON src.id = a.source_id").
		Where("a.id = ? AND src.user_id = ?", articleID, userID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrArticleMissing
	}
	return nil
}

// upsertState 更新或创建用户文章状态，mut 只改目标字段，另一状态字段保持不变
func (s *ArticleService) upsertState(userID, articleID int64, mut func(*model.UserArticleState)) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var st model.UserArticleState
		err := tx.Where("user_id = ? AND article_id = ?", userID, articleID).Take(&st).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			st = model.UserArticleState{UserID: userID, ArticleID: articleID}
			mut(&st)
			return tx.Create(&st).Error
		}
		if err != nil {
			return err
		}
		mut(&st)
		return tx.Save(&st).Error
	})
}

// buildListQuery 构造带归属/筛选/搜索条件的查询（不含 select/order/分页）
func (s *ArticleService) buildListQuery(userID int64, filter, search string, sourceID int64) *gorm.DB {
	q := s.baseScopedQuery(userID)

	switch filter {
	case "", FilterAll:
		// 无额外条件
	case FilterUnread:
		q = q.Where("COALESCE(st.is_read, 0) = 0")
	case FilterRead:
		q = q.Where("st.is_read = 1")
	case FilterFavorited:
		q = q.Where("st.is_favorited = 1")
	case FilterToday:
		now := time.Now().In(chinaZone)
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, chinaZone)
		q = q.Where("a.published_at >= ?", startOfDay.UTC())
	default:
		// 未知 filter 按 all 处理
	}

	if kw := strings.TrimSpace(search); kw != "" {
		like := "%" + escapeLike(kw) + "%"
		q = q.Where("(a.title LIKE ? ESCAPE '\\' OR a.summary LIKE ? ESCAPE '\\')", like, like)
	}
	if sourceID > 0 {
		q = q.Where("a.source_id = ?", sourceID)
	}
	return q
}

// baseScopedQuery 文章 join 订阅源并限定当前用户的基础查询
func (s *ArticleService) baseScopedQuery(userID int64) *gorm.DB {
	return s.db.Table("articles a").
		Joins("JOIN sources src ON src.id = a.source_id AND src.user_id = ?", userID).
		Joins("LEFT JOIN user_article_states st ON st.article_id = a.id AND st.user_id = ?", userID)
}

func toListItem(r articleRow) ArticleListItem {
	return ArticleListItem{
		ID:          r.ID,
		SourceID:    r.SourceID,
		SourceName:  r.SourceName,
		Title:       r.Title,
		Summary:     r.Summary,
		Author:      r.Author,
		Link:        r.Link,
		PublishedAt: r.PublishedAt.Unix(),
		IsRead:      r.IsRead,
		IsFavorited: r.IsFavorited,
	}
}

// escapeLike 转义 LIKE 通配符，配合 ESCAPE '\' 使用
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// RewriteImageProxy 把正文中所有 http(s) 图片地址重写为图片代理地址，
// 数据库保留原始 URL，仅在响应时改写
func RewriteImageProxy(content, proxyPrefix string) string {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return content
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for i, attr := range n.Attr {
				if attr.Key == "src" && strings.TrimSpace(attr.Val) != "" {
					u := strings.TrimSpace(attr.Val)
					if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
						n.Attr[i].Val = proxyPrefix + url.QueryEscape(u)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return content
	}
	rendered := buf.String()
	if start := strings.Index(rendered, "<body>"); start >= 0 {
		start += len("<body>")
		if end := strings.LastIndex(rendered, "</body>"); end > start {
			return rendered[start:end]
		}
	}
	return rendered
}
