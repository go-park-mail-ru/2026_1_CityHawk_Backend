package postgres

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"cityhawk/backend/internal/platform/httpx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.Tracer = postgresTracer{}
	return pgxpool.NewWithConfig(ctx, cfg)
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
