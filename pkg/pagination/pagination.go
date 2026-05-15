package pagination

import (
	"github.com/gin-gonic/gin"
	"math"
	"strconv"
)

type Params struct{ Page, Size, Offset int }

func FromCtx(c *gin.Context) Params {
	page := parseInt(c.Query("page"), 1)
	size := parseInt(c.Query("page_size"), 20)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return Params{Page: page, Size: size, Offset: (page - 1) * size}
}

type Page[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func New[T any](data []T, total int64, p Params) Page[T] {
	tp := int(math.Ceil(float64(total) / float64(p.Size)))
	if tp < 1 {
		tp = 1
	}
	return Page[T]{Data: data, Total: total, Page: p.Page, PageSize: p.Size, TotalPages: tp, HasNext: p.Page < tp, HasPrev: p.Page > 1}
}

func parseInt(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
