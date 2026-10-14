package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/service"
)

// GET /api/article/list
func (h *Handler) articleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))

	var sourceID int64
	if raw := c.Query("source_id"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			sourceID = v
		}
	}

	result, err := h.article.List(service.ArticleListParams{
		UserID:   uid(c),
		Filter:   c.DefaultQuery("filter", service.FilterAll),
		Search:   c.Query("search"),
		SourceID: sourceID,
		Page:     page,
		Size:     size,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, result)
}

// GET /api/article/detail/:id
func (h *Handler) articleDetail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	detail, err := h.article.Detail(uid(c), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, detail)
}

// POST /api/article/read/:id
func (h *Handler) articleMarkRead(c *gin.Context) {
	h.markRead(c, true)
}

// POST /api/article/unread/:id
func (h *Handler) articleMarkUnread(c *gin.Context) {
	h.markRead(c, false)
}

func (h *Handler) markRead(c *gin.Context, read bool) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.article.SetRead(uid(c), id, read); err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, gin.H{"id": id, "is_read": read})
}

// POST /api/article/favorite/:id
func (h *Handler) articleFavorite(c *gin.Context) {
	h.markFavorite(c, true)
}

// POST /api/article/unfavorite/:id
func (h *Handler) articleUnfavorite(c *gin.Context) {
	h.markFavorite(c, false)
}

func (h *Handler) markFavorite(c *gin.Context, favorited bool) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.article.SetFavorite(uid(c), id, favorited); err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, gin.H{"id": id, "is_favorited": favorited})
}
