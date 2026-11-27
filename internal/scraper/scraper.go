package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/exemayj/scraper/internal/parser"
)

type Scraper struct {
	Client  *http.Client
	Workers int
}

type Job struct {
	URL string
}

func New(workers int) *Scraper {
	return &Scraper{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Workers: workers,
	}
}

// Run скачивает страницы и собирает ссылки.
func (s *Scraper) Run(ctx context.Context, urls []string) ([]parser.Article, error) {
	jobs := make(chan string, len(urls))
	results := make(chan []parser.Article, len(urls))

	var wg sync.WaitGroup

	// Запускаем воркеров
	for i := 0; i < s.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				articles, err := s.fetchAndParse(ctx, url)
				if err != nil {
					fmt.Printf("error on %s: %v\n", url, err)
					continue
				}
				results <- articles
			}
		}()
	}

	// Кидаем задачи
	go func() {
		for _, u := range urls {
			jobs <- u
		}
		close(jobs)
	}()

	// Ждём воркеров и закрываем результаты
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем
	var all []parser.Article
	for a := range results {
		all = append(all, a...)
	}

	return all, nil
}

func (s *Scraper) fetchAndParse(ctx context.Context, url string) ([]parser.Article, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Уважаем сайты: ставим User-Agent
	req.Header.Set("User-Agent", "GoScraper/1.0 (+https://github.com/exemayj/scraper)")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, err
	}

	return parser.ExtractLinks(string(body), url), nil
}
