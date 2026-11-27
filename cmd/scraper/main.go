package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/exemayj/scraper/internal/scraper"
)

func main() {
	url := flag.String("url", "", "URL to scrape")
	workers := flag.Int("workers", 4, "number of concurrent workers")
	out := flag.String("out", "articles.json", "output file")
	flag.Parse()

	if *url == "" {
		log.Fatal("--url is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s := scraper.New(*workers)
	articles, err := s.Run(ctx, []string{*url})
	if err != nil {
		log.Fatalf("scrape: %v", err)
	}

	// Дедупликация по URL
	seen := make(map[string]bool)
	unique := make([]interface{}, 0, len(articles))
	for _, a := range articles {
		if seen[a.URL] {
			continue
		}
		seen[a.URL] = true
		unique = append(unique, a)
	}

	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("create file: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(unique); err != nil {
		log.Fatalf("encode: %v", err)
	}

	log.Printf("saved %d unique articles to %s", len(unique), *out)
}
