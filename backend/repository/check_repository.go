package repository

import (
	"context"
	"database/sql"
	"github.com/koloo91/monhttp/model"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	selectChecksByServiceIdAndCreatedAtStatement *sql.Stmt
	selectAverageLatencyStatement                *sql.Stmt
)

func prepareCheckStatements() {
	var err error

	selectChecksByServiceIdAndCreatedAtStatement, err = db.Prepare(`SELECT id, 
					   latency_in_ms, 
					   is_failure, 
					   created_at FROM "check"
					   WHERE service_id = $1 AND created_at >= $2 AND created_at <= $3
						ORDER BY created_at DESC;`)
	if err != nil {
		log.Fatal(err)
	}

	selectAverageLatencyStatement, err = db.Prepare(`SELECT COALESCE(ROUND(AVG(latency_in_ms)), 0)
															FROM "check"
															WHERE service_id = $1
															  AND created_at >= $2
															  AND created_at <= $3;`)
	if err != nil {
		log.Fatal(err)
	}
}

func InsertCheck(ctx context.Context, tx *sql.Tx, check model.Check) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO "check" (id, service_id, latency_in_ms, is_failure, created_at) 
											VALUES ($1, $2, $3, $4, $5)`,
		check.Id, check.ServiceId, check.LatencyInMs, check.IsFailure, check.CreatedAt); err != nil {
		return err
	}
	return nil
}

func SelectChecks(ctx context.Context, serviceId string, from, to time.Time) ([]model.Check, error) {
	rows, err := selectChecksByServiceIdAndCreatedAtStatement.QueryContext(ctx, serviceId, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id string
	var latencyInMs int64
	var isFailure bool
	var createdAt time.Time

	result := make([]model.Check, 0)

	for rows.Next() {
		if err := rows.Scan(&id, &latencyInMs, &isFailure, &createdAt); err != nil {
			return nil, err
		}

		result = append(result, model.Check{
			Id:          id,
			ServiceId:   serviceId,
			LatencyInMs: latencyInMs,
			IsFailure:   isFailure,
			CreatedAt:   createdAt,
		})
	}

	return result, nil
}

func SelectAverageLatency(ctx context.Context, serviceId string, from time.Time, to time.Time) (int, error) {
	row := selectAverageLatencyStatement.QueryRowContext(ctx, serviceId, from, to)

	var avg int
	if err := row.Scan(&avg); err != nil {
		return 0, err
	}
	return avg, nil
}
