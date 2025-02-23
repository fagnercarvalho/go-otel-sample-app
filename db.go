package main

import (
	"context"
	"database/sql"

	"github.com/XSAM/otelsql"
	otelpyroscope "github.com/grafana/otel-profiling-go"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

func NewDB(serviceName string, tracerProvider trace.TracerProvider, meterProvider metric.MeterProvider) (DB, error) {
	sqlDB, err := otelsql.Open(
		"sqlite",
		":memory:",
		otelsql.WithAttributes(
			semconv.DBSystemSqlite,
			semconv.ServiceNameKey.String(serviceName),
		),
		otelsql.WithTracerProvider(otelpyroscope.NewTracerProvider(tracerProvider)),
		otelsql.WithMeterProvider(meterProvider),
	)
	if err != nil {
		return DB{}, err
	}

	err = otelsql.RegisterDBStatsMetrics(sqlDB, otelsql.WithAttributes(
		semconv.DBSystemSqlite,
	))
	if err != nil {
		return DB{}, err
	}

	db := DB{db: sqlDB}

	return db, db.initialize()
}

func (handler *DB) AddTodo(ctx context.Context, task string, done bool) error {
	query := `INSERT INTO todos (task, done) VALUES (?, ?)`

	_, err := handler.db.ExecContext(ctx, query, task, done)
	return err
}

func (handler *DB) FetchAllTodos(ctx context.Context) ([]Todo, error) {
	rows, err := handler.db.QueryContext(ctx, `SELECT id, task, done FROM todos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		var done bool
		err := rows.Scan(&t.ID, &t.Task, &done)
		if err != nil {
			return nil, err
		}

		t.Done = done
		todos = append(todos, t)
	}

	return todos, nil
}

func (handler *DB) Close() error {
	return handler.db.Close()
}

func (handler *DB) initialize() error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task TEXT NOT NULL,
		done BOOLEAN NOT NULL CHECK (done IN (0, 1))
	)`

	_, err := handler.db.Exec(query)
	return err
}
