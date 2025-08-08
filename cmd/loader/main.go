package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"
)

func main() {
	var (
		baseURL  = flag.String("base-url", "http://localhost:8080", "service base URL")
		mode     = flag.String("mode", "single", "mode: single|range")
		bannerID = flag.Int("banner", 1, "banner id (for single mode)")
		minID    = flag.Int("min", 1, "min banner id (for range mode)")
		maxID    = flag.Int("max", 100, "max banner id (for range mode)")
		rps      = flag.Int("rps", 500, "requests per second")
		timeout  = flag.Duration("timeout", 3*time.Second, "http client timeout")
		duration = flag.Duration("duration", 0, "test duration (e.g. 30s). 0 = run until Ctrl+C")
	)
	flag.Parse()

	if *rps <= 0 {
		log.Fatalf("rps must be > 0")
	}
	if *mode != "single" && *mode != "range" {
		log.Fatalf("mode must be 'single' or 'range'")
	}
	if *mode == "range" && *minID > *maxID {
		log.Fatalf("min must be <= max")
	}

	client := &http.Client{
		Timeout: *timeout,
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Контекст отмены: по Ctrl+C или по duration
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if *duration > 0 {
		go func() {
			timer := time.NewTimer(*duration)
			defer timer.Stop()
			<-timer.C
			cancel()
		}()
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
	}()

	// Вычисляем число горутин по формуле goroutines = max(1, rps/10)
	n := *rps / 10
	if n < 1 {
		n = 1
	}
	// Равномерно распределяем целевой rps между горутинами
	per := *rps / n
	rem := *rps % n

	trimBase := strings.TrimRight(*baseURL, "/")
	pickBanner := makePicker(*mode, *bannerID, *minID, *maxID)

	var okCnt, errCnt atomic.Uint64

	// Репортинг QPS
	go func() {
		var prevOK, prevErr uint64
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				okNow := okCnt.Load()
				errNow := errCnt.Load()
				fmt.Printf("OK/s=%d ERR/s=%d  totalOK=%d totalERR=%d\n",
					okNow-prevOK, errNow-prevErr, okNow, errNow)
				prevOK, prevErr = okNow, errNow
			}
		}
	}()

	// Запуск воркеров
	for i := 0; i < n; i++ {
		rate := per
		if i < rem {
			rate++ // распределим остаток по первым воркерам
		}
		go worker(ctx, client, trimBase, pickBanner, rate, &okCnt, &errCnt)
	}

	<-ctx.Done()
	// Дать воркерам чуть времени корректно завершиться
	time.Sleep(200 * time.Millisecond)
}

func makePicker(mode string, singleID, minID, maxID int) func() int {
	if mode == "single" {
		return func() int { return singleID }
	}
	// range — выбрасываем случайный id в [min,max]
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return func() int { return minID + r.Intn(maxID-minID+1) }
}

func worker(ctx context.Context, client *http.Client, base string, pick func() int, rps int, okCnt, errCnt *atomic.Uint64) {
	if rps <= 0 {
		return
	}
	interval := time.Second / time.Duration(rps)
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			id := pick()
			url := fmt.Sprintf("%s/counter/%d", base, id)

			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			req.Header.Set("Cache-Control", "no-store")
			resp, err := client.Do(req)
			if err != nil {
				errCnt.Add(1)
				continue
			}
			// Ожидаем 204, но не критично — считаем любой 2xx как успех
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				okCnt.Add(1)
			} else {
				errCnt.Add(1)
			}
			_ = resp.Body.Close()
		}
	}
}
