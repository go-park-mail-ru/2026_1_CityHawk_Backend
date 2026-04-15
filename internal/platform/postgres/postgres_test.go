package postgres

import (
	"context"
	"errors"
	"testing"

	"cityhawk/backend/internal/platform/httpx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCompactSQLAndFormatArgs(t *testing.T) {
	if got := compactSQL(" SELECT  * \n FROM users \t WHERE id = $1 "); got != "SELECT * FROM users WHERE id = $1" {
		t.Fatalf("compactSQL() = %q", got)
	}
	if got := formatArgs(nil); got != "[]" {
		t.Fatalf("formatArgs(nil) = %q", got)
	}
	if got := formatArgs([]any{"a", 1}); got != "[a 1]" {
		t.Fatalf("formatArgs() = %q", got)
	}
}

func TestPostgresTracerTraceQueryStartAndEnd(t *testing.T) {
	tracer := postgresTracer{}
	ctx := context.WithValue(context.Background(), httpx.RequestIDContextKey, "req-1")
	ctx = tracer.TraceQueryStart(ctx, nil, pgx.TraceQueryStartData{
		SQL:  "SELECT *\nFROM users WHERE id = $1",
		Args: []any{"user-1"},
	})

	data, ok := ctx.Value(queryStartContextKey).(queryLogData)
	if !ok || data.sql != "SELECT * FROM users WHERE id = $1" || data.args != "[user-1]" {
		t.Fatalf("unexpected query log data: %+v", data)
	}

	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
		CommandTag: pgconn.CommandTag{},
		Err:        errors.New("boom"),
	})
}
