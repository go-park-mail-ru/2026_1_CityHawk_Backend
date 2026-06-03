package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"cityhawk/backend/internal/platform/httpx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

func TestApplyPoolConfig(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/cityhawk?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	applyPoolConfig(cfg, PoolConfig{
		ApplicationName:  "cityhawk-test",
		MaxConns:         12,
		MinConns:         2,
		MaxConnLifetime:  20 * time.Minute,
		MaxConnIdleTime:  4 * time.Minute,
		StatementTimeout: 3 * time.Second,
		LockTimeout:      750 * time.Millisecond,
	})

	if cfg.MaxConns != 12 || cfg.MinConns != 2 {
		t.Fatalf("unexpected pool sizes: max=%d min=%d", cfg.MaxConns, cfg.MinConns)
	}
	if cfg.MaxConnLifetime != 20*time.Minute || cfg.MaxConnIdleTime != 4*time.Minute {
		t.Fatalf("unexpected pool lifetimes: %+v", cfg)
	}
	if got := cfg.ConnConfig.RuntimeParams["application_name"]; got != "cityhawk-test" {
		t.Fatalf("application_name = %q", got)
	}
	if got := cfg.ConnConfig.RuntimeParams["statement_timeout"]; got != "3000ms" {
		t.Fatalf("statement_timeout = %q", got)
	}
	if got := cfg.ConnConfig.RuntimeParams["lock_timeout"]; got != "750ms" {
		t.Fatalf("lock_timeout = %q", got)
	}
}
