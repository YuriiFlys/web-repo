package httpx

import (
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type queryTestModel struct {
	ID        uint
	Title     string
	CreatedAt string
}

func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(noopDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open gorm dry run db: %v", err)
	}

	return db
}

type noopDialector struct{}

func (noopDialector) Name() string { return "noop" }

func (noopDialector) Initialize(*gorm.DB) error { return nil }

func (noopDialector) Migrator(*gorm.DB) gorm.Migrator { return nil }

func (noopDialector) DataTypeOf(*schema.Field) string { return "" }

func (noopDialector) DefaultValueOf(*schema.Field) clause.Expression { return clause.Expr{} }

func (noopDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v any) {
	writer.WriteByte('?')
}

func (noopDialector) QuoteTo(writer clause.Writer, s string) {
	writer.WriteString(s)
}

func (noopDialector) Explain(sql string, vars ...any) string {
	return sql
}

func TestParseListParams(t *testing.T) {
	params := ParseListParams("0", "101", "-title, createdAt")
	if params.Page != 1 || params.PageSize != 100 {
		t.Fatalf("params = %+v", params)
	}
	if len(params.Sort) != 2 || params.Sort[0].Field != "title" || !params.Sort[0].Desc || params.Sort[1].Field != "createdAt" || params.Sort[1].Desc {
		t.Fatalf("sort = %+v", params.Sort)
	}
}

func TestParseListParamsNormalizesSmallValuesAndEmptySortParts(t *testing.T) {
	params := ParseListParams("", "0", " , -title")
	if params.Page != 1 || params.PageSize != 20 {
		t.Fatalf("params = %+v", params)
	}
	if len(params.Sort) != 1 || params.Sort[0].Field != "title" || !params.Sort[0].Desc {
		t.Fatalf("sort = %+v", params.Sort)
	}
}

func TestIsLast(t *testing.T) {
	if !IsLast(0, ListParams{Page: 1, PageSize: 20}) {
		t.Fatal("expected empty result to be last")
	}
	if IsLast(41, ListParams{Page: 2, PageSize: 20}) {
		t.Fatal("did not expect page 2/20 on total 41 to be last")
	}
	if !IsLast(40, ListParams{Page: 2, PageSize: 20}) {
		t.Fatal("expected exact end to be last")
	}
}

func TestSplitCommaAndAtoiDefault(t *testing.T) {
	if got := splitComma(" "); got != nil {
		t.Fatalf("splitComma blank = %#v", got)
	}
	got := splitComma("a,b")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("splitComma = %#v", got)
	}
	if got := atoiDefault("bad", 3); got != 3 {
		t.Fatalf("atoiDefault bad = %d", got)
	}
	if got := atoiDefault("7", 3); got != 7 {
		t.Fatalf("atoiDefault = %d", got)
	}
}

func TestApplyPagination(t *testing.T) {
	db := newDryRunDB(t)

	tx := ApplyPagination(db.Model(&queryTestModel{}), ListParams{Page: 3, PageSize: 10})
	limitClause, ok := tx.Statement.Clauses["LIMIT"]
	if !ok {
		t.Fatal("expected LIMIT clause")
	}
	limitExpr, ok := limitClause.Expression.(clause.Limit)
	if !ok {
		t.Fatalf("unexpected limit expression: %#v", limitClause.Expression)
	}
	if limitExpr.Limit == nil || *limitExpr.Limit != 10 || limitExpr.Offset != 20 {
		t.Fatalf("limit = %#v", limitExpr)
	}
}

func TestApplySorting(t *testing.T) {
	db := newDryRunDB(t)
	allowed := map[string]string{
		"title":     "title",
		"createdAt": "created_at",
	}

	t.Run("applies valid sorts", func(t *testing.T) {
		tx := ApplySorting(db.Model(&queryTestModel{}), allowed, ListParams{
			Sort: []SortField{
				{Field: "title", Desc: true},
				{Field: "createdAt"},
			},
		}, "created_at DESC")
		orderClause, ok := tx.Statement.Clauses["ORDER BY"]
		if !ok {
			t.Fatal("expected ORDER BY clause")
		}
		orderExpr, ok := orderClause.Expression.(clause.OrderBy)
		if !ok {
			t.Fatalf("unexpected order expression: %#v", orderClause.Expression)
		}
		if len(orderExpr.Columns) != 2 || orderExpr.Columns[0].Column.Name != "title DESC" || !orderExpr.Columns[0].Column.Raw || orderExpr.Columns[1].Column.Name != "created_at ASC" || !orderExpr.Columns[1].Column.Raw {
			t.Fatalf("order columns = %#v", orderExpr.Columns)
		}
	})

	t.Run("falls back to default order for invalid sort", func(t *testing.T) {
		tx := ApplySorting(db.Model(&queryTestModel{}), allowed, ListParams{
			Sort: []SortField{{Field: "unknown", Desc: true}},
		}, "created_at DESC")
		orderClause, ok := tx.Statement.Clauses["ORDER BY"]
		if !ok {
			t.Fatal("expected ORDER BY clause")
		}
		orderExpr, ok := orderClause.Expression.(clause.OrderBy)
		if !ok {
			t.Fatalf("unexpected order expression: %#v", orderClause.Expression)
		}
		if len(orderExpr.Columns) != 1 || orderExpr.Columns[0].Column.Name != "created_at DESC" || !orderExpr.Columns[0].Column.Raw {
			t.Fatalf("order columns = %#v", orderExpr.Columns)
		}
	})

	t.Run("skips default order when empty", func(t *testing.T) {
		tx := ApplySorting(db.Model(&queryTestModel{}), allowed, ListParams{
			Sort: []SortField{{Field: "unknown"}},
		}, "")
		if _, ok := tx.Statement.Clauses["ORDER BY"]; ok {
			t.Fatal("did not expect ORDER BY clause")
		}
	})
}
