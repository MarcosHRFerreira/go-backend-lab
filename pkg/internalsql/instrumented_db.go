package internalsql

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type QueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Database interface {
	QueryExecutor
	PingContext(ctx context.Context) error
	Close() error
}

type QueryObserver interface {
	ObserveQuery(query string, duration time.Duration)
}

type instrumentedDB struct {
	raw      *sql.DB
	observer QueryObserver
	tracer   trace.Tracer
}

func NewInstrumentedDB(raw *sql.DB, observer QueryObserver, tracer trace.Tracer) Database {
	if observer == nil && tracer == nil {
		return raw
	}

	return &instrumentedDB{
		raw:      raw,
		observer: observer,
		tracer:   tracer,
	}
}

func (db *instrumentedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	ctx, span := db.startSpan(ctx, query)
	defer span.End()

	startedAt := time.Now()
	result, err := db.raw.ExecContext(ctx, query, args...)
	db.observeQuery(query, time.Since(startedAt))
	db.finishSpan(span, err)
	return result, err
}

func (db *instrumentedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	ctx, span := db.startSpan(ctx, query)
	defer span.End()

	startedAt := time.Now()
	rows, err := db.raw.QueryContext(ctx, query, args...)
	db.observeQuery(query, time.Since(startedAt))
	db.finishSpan(span, err)
	return rows, err
}

func (db *instrumentedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	ctx, span := db.startSpan(ctx, query)
	defer span.End()

	startedAt := time.Now()
	row := db.raw.QueryRowContext(ctx, query, args...)
	db.observeQuery(query, time.Since(startedAt))
	return row
}

func (db *instrumentedDB) PingContext(ctx context.Context) error {
	return db.raw.PingContext(ctx)
}

func (db *instrumentedDB) Close() error {
	return db.raw.Close()
}

func (db *instrumentedDB) observeQuery(query string, duration time.Duration) {
	if db.observer == nil {
		return
	}

	db.observer.ObserveQuery(query, duration)
}

func (db *instrumentedDB) startSpan(ctx context.Context, query string) (context.Context, trace.Span) {
	if db.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	operation, table := queryMetadata(query)
	return db.tracer.Start(ctx, "db."+operation, trace.WithAttributes(
		attribute.String("db.system", "mysql"),
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	))
}

func (db *instrumentedDB) finishSpan(span trace.Span, err error) {
	if span == nil {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func queryMetadata(query string) (string, string) {
	normalized := strings.Join(strings.Fields(strings.ToLower(query)), " ")
	if normalized == "" {
		return "query", "unknown"
	}

	switch {
	case strings.HasPrefix(normalized, "select"):
		return "select", tableAfterKeyword(normalized, " from ")
	case strings.HasPrefix(normalized, "insert"):
		return "insert", tableAfterKeyword(normalized, " into ")
	case strings.HasPrefix(normalized, "update"):
		return "update", tableAfterPrefix(normalized, "update ")
	case strings.HasPrefix(normalized, "delete"):
		return "delete", tableAfterKeyword(normalized, " from ")
	default:
		return "query", "unknown"
	}
}

func tableAfterKeyword(query string, keyword string) string {
	index := strings.Index(query, keyword)
	if index == -1 {
		return "unknown"
	}

	return cleanTableName(query[index+len(keyword):])
}

func tableAfterPrefix(query string, prefix string) string {
	if !strings.HasPrefix(query, prefix) {
		return "unknown"
	}

	return cleanTableName(strings.TrimPrefix(query, prefix))
}

func cleanTableName(fragment string) string {
	fields := strings.Fields(fragment)
	if len(fields) == 0 {
		return "unknown"
	}

	table := strings.Trim(fields[0], "` ,;")
	if table == "" {
		return "unknown"
	}

	return table
}
