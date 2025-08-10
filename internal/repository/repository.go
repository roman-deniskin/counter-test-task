package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type ClickRow struct {
	BannerID int64
	TsMinute time.Time
	Count    uint64
}

// UpsertBatch добавляет агрегаты поминутно: count = count + EXCLUDED.count
func (r *Repository) UpsertBatch(ctx context.Context, rows []ClickRow) error {
	if len(rows) == 0 {
		return nil
	}
	// Готовим VALUES ($1,$2,$3),($4,$5,$6)...
	var (
		sb    strings.Builder
		args  = make([]any, 0, len(rows)*3)
		place = 1
	)
	sb.WriteString("INSERT INTO banner_clicks (banner_id, ts_minute, count) VALUES ")
	for i, row := range rows {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(fmt.Sprintf("($%d,$%d,$%d)", place, place+1, place+2))
		args = append(args, row.BannerID, row.TsMinute, row.Count)
		place += 3
	}
	sb.WriteString(" ON CONFLICT (banner_id, ts_minute) DO UPDATE SET ")
	sb.WriteString(" count = banner_clicks.count + EXCLUDED.count")

	// Один Exec в транзакции на всякий
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type StatsRow struct {
	Ts    time.Time
	Count int
}

func (r *Repository) SelectStatsFromDB(ctx context.Context, bannerID int64, from, to time.Time) ([]StatsRow, error) {
	const q = `
SELECT ts_minute, count
FROM banner_clicks
WHERE banner_id = $1
  AND ts_minute >= $2
  AND ts_minute <  $3
ORDER BY ts_minute;
`
	rows, err := r.db.QueryContext(ctx, q, bannerID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]StatsRow, 0, 128)
	for rows.Next() {
		var sr StatsRow
		if err := rows.Scan(&sr.Ts, &sr.Count); err != nil {
			return nil, err
		}
		out = append(out, sr)
	}
	return out, rows.Err()
}
