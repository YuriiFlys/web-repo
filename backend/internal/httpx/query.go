package httpx

import (
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type ListParams struct {
	Page     int
	PageSize int
	Sort     []SortField
}

type SortField struct {
	Field string
	Desc  bool
}

func ParseListParams(pageStr, sizeStr, sortStr string) ListParams {
	page := atoiDefault(pageStr, 1)
	size := atoiDefault(sizeStr, 20)

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	var sorts []SortField
	for _, part := range splitComma(sortStr) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		desc := strings.HasPrefix(part, "-")
		field := strings.TrimPrefix(part, "-")
		sorts = append(sorts, SortField{Field: field, Desc: desc})
	}

	return ListParams{Page: page, PageSize: size, Sort: sorts}
}

func ApplyPagination(db *gorm.DB, p ListParams) *gorm.DB {
	offset := (p.Page - 1) * p.PageSize
	return db.Offset(offset).Limit(p.PageSize)
}

func ApplySorting(db *gorm.DB, allowed map[string]string, p ListParams, defaultOrder string) *gorm.DB {
	applied := false
	for _, s := range p.Sort {
		col, ok := allowed[s.Field]
		if !ok {
			continue
		}
		order := col
		if s.Desc {
			order += " DESC"
		} else {
			order += " ASC"
		}
		db = db.Order(order)
		applied = true
	}
	if !applied && defaultOrder != "" {
		db = db.Order(defaultOrder)
	}
	return db
}

func IsLast(total int64, p ListParams) bool {
	if total == 0 {
		return true
	}
	end := int64(p.Page * p.PageSize)
	return end >= total
}

func splitComma(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
