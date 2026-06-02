package postgres

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"cityhawk/backend/internal/platform/httpx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConfig struct {
	ApplicationName  string
	MaxConns         int
	MinConns         int
	MaxConnLifetime  time.Duration
	MaxConnIdleTime  time.Duration
	StatementTimeout time.Duration
	LockTimeout      time.Duration
}

func NewPool(ctx context.Context, dsn string, options ...PoolConfig) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.Tracer = postgresTracer{}
	if len(options) > 0 {
		applyPoolConfig(cfg, options[0])
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

func applyPoolConfig(cfg *pgxpool.Config, option PoolConfig) {
	if option.MaxConns > 0 {
		cfg.MaxConns = int32(option.MaxConns)
	}
	if option.MinConns > 0 {
		cfg.MinConns = int32(option.MinConns)
	}
	if option.MaxConnLifetime > 0 {
		cfg.MaxConnLifetime = option.MaxConnLifetime
	}
	if option.MaxConnIdleTime > 0 {
		cfg.MaxConnIdleTime = option.MaxConnIdleTime
	}

	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	if strings.TrimSpace(option.ApplicationName) != "" {
		cfg.ConnConfig.RuntimeParams["application_name"] = strings.TrimSpace(option.ApplicationName)
	}
	if option.StatementTimeout > 0 {
		cfg.ConnConfig.RuntimeParams["statement_timeout"] = formatPostgresDuration(option.StatementTimeout)
	}
	if option.LockTimeout > 0 {
		cfg.ConnConfig.RuntimeParams["lock_timeout"] = formatPostgresDuration(option.LockTimeout)
	}
}

func formatPostgresDuration(duration time.Duration) string {
	return strconv.FormatInt(duration.Milliseconds(), 10) + "ms"
}

type postgresTracer struct{}

type tracerContextKey string

const queryStartContextKey tracerContextKey = "postgresQueryStart"

func (postgresTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, queryStartContextKey, queryLogData{
		startedAt: time.Now(),
		sql:       compactSQL(data.SQL),
		args:      formatArgs(data.Args),
	})
}

func (postgresTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	queryData, _ := ctx.Value(queryStartContextKey).(queryLogData)
	requestID, _ := ctx.Value(httpx.RequestIDContextKey).(string)

	log.Printf(
		"db request_id=%s sql=%q args=%s duration=%s command_tag=%s err=%v",
		requestID,
		queryData.sql,
		queryData.args,
		time.Since(queryData.startedAt).Round(time.Millisecond),
		data.CommandTag.String(),
		data.Err,
	)
}

type queryLogData struct {
	startedAt time.Time
	sql       string
	args      string
}

func compactSQL(sql string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
}

func formatArgs(args []any) string {
	if len(args) == 0 {
		return "[]"
	}

	return fmt.Sprintf("%v", args)
}
