package service

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sort"
	"time"

	"counter-test-task/internal/repository"
)

type Service struct {
	DB            *sql.DB
	repo          *repository.Repository
	agg           *aggregator
	flushInterval time.Duration
	stopCh        chan struct{}
}

func New(db *sql.DB) *Service {
	s := &Service{
		DB:            db,
		repo:          repository.New(db),
		agg:           newAggregator(),
		flushInterval: time.Second,
		stopCh:        make(chan struct{}),
	}
	go s.runFlusher()
	go s.trapSignals()
	return s
}

func (s *Service) PingDB() error { return s.DB.Ping() }

func (s *Service) IncClick(bannerID int64) error {
	s.agg.Inc(bannerID, time.Now().UTC())
	return nil
}

type StatPoint struct {
	TS time.Time
	V  int
}

func (s *Service) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]StatPoint, error) {
	from = from.UTC().Truncate(time.Minute)
	to = to.UTC().Truncate(time.Minute)

	// Данные из Postgres
	rows, err := s.repo.SelectStatsFromDB(ctx, bannerID, from, to)
	if err != nil {
		return nil, err
	}

	// Данные из текущей map где хранятся значения не попавшие пока в БД
	pending := s.agg.SelectStatsFromMap(bannerID, from, to)

	m := make(map[time.Time]int, len(rows)+len(pending))
	for _, r := range rows {
		m[r.Ts] += r.Count
	}
	for ts, dv := range pending {
		if dv > 0 {
			m[ts] += int(dv)
		}
	}

	tsList := make([]time.Time, 0, len(m))
	for ts, v := range m {
		if v > 0 {
			tsList = append(tsList, ts)
		}
	}

	sort.Slice(tsList, func(i, j int) bool { return tsList[i].Before(tsList[j]) })

	out := make([]StatPoint, 0, len(tsList))
	for _, ts := range tsList {
		out = append(out, StatPoint{TS: ts, V: m[ts]})
	}
	return out, nil
}

func (s *Service) runFlusher() {
	t := time.NewTicker(s.flushInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s.flushOnce(context.Background())
		case <-s.stopCh:
			s.flushOnce(context.Background())
			return
		}
	}
}

func (s *Service) flushOnce(ctx context.Context) {
	batch := s.agg.SnapshotAndClear()
	if len(batch) == 0 {
		return
	}
	if err := s.repo.UpsertBatch(ctx, batch); err != nil {
		log.Printf("flush error, requeue: %v", err)
		s.agg.Requeue(batch)
	}
}

func (s *Service) Close() {
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
}

func (s *Service) trapSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	s.Close()
}
