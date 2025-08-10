package service

import (
	"sync"
	"time"

	"counter-test-task/internal/repository"
)

type aggKey struct {
	id int64
	tm int64 // minute.Unix()
}

type aggVal struct {
	cnt uint64
}

type aggregator struct {
	mu sync.Mutex
	m  map[aggKey]aggVal
}

func newAggregator() *aggregator {
	return &aggregator{m: make(map[aggKey]aggVal, 1024)}
}

func (a *aggregator) Inc(bannerID int64, t time.Time) {
	tu := t.UTC()
	minTu := tu.Truncate(time.Minute).Unix()
	k := aggKey{id: bannerID, tm: minTu}

	a.mu.Lock()
	v := a.m[k]
	v.cnt++
	a.m[k] = v
	a.mu.Unlock()
}

func (a *aggregator) SnapshotAndClear() []repository.ClickRow {
	a.mu.Lock()
	snap := a.m
	a.m = make(map[aggKey]aggVal, len(snap))
	a.mu.Unlock()

	if len(snap) == 0 {
		return nil
	}
	out := make([]repository.ClickRow, 0, len(snap))
	for k, v := range snap {
		out = append(out, repository.ClickRow{
			BannerID: k.id,
			TsMinute: time.Unix(k.tm, 0).UTC(),
			Count:    v.cnt,
		})
	}
	return out
}

func (a *aggregator) Requeue(rows []repository.ClickRow) {
	if len(rows) == 0 {
		return
	}
	a.mu.Lock()
	for _, r := range rows {
		k := aggKey{id: r.BannerID, tm: r.TsMinute.UTC().Truncate(time.Minute).Unix()}
		v := a.m[k]
		v.cnt += r.Count
		a.m[k] = v
	}
	a.mu.Unlock()
}

func (a *aggregator) SelectStatsFromMap(bannerID int64, from, to time.Time) map[time.Time]uint64 {
	fromMin := from.UTC().Truncate(time.Minute).Unix()
	toMin := to.UTC().Truncate(time.Minute).Unix() // правая граница не включается

	res := make(map[time.Time]uint64)
	a.mu.Lock()
	for k, v := range a.m {
		if k.id != bannerID {
			continue
		}
		if k.tm >= fromMin && k.tm < toMin {
			res[time.Unix(k.tm, 0).UTC()] += v.cnt
		}
	}
	a.mu.Unlock()
	return res
}
