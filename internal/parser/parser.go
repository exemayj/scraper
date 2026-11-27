package parser

import (
	"regexp"
	"strings"
)

// Article — то, что мы сохраняем.
type Article struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// <a ... href="URL" ...>Title</a>
var linkRe = regexp.MustCompile(`(?i)<a\s+[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>`)

var tagRe = regexp.MustCompile(`<[^>]+>`)

// ExtractLinks парсит HTML и возвращает ссылки с текстом.
func ExtractLinks(html string, baseURL string) []Article {
	matches := linkRe.FindAllStringSubmatch(html, -1)
	out := make([]Article, 0, len(matches))

	for _, m := range matches {
		href := strings.TrimSpace(m[1])
		title := strings.TrimSpace(tagRe.ReplaceAllString(m[2], ""))
		title = strings.Join(strings.Fields(title), " ") // убираем лишние пробелы

		if href == "" || title == "" {
			continue
		}
		// отсекаем якоря и javascript
		if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
			continue
		}

		// простейшая нормализация относительных ссылок
		if strings.HasPrefix(href, "/") && baseURL != "" {
			href = strings.TrimSuffix(baseURL, "/") + href
		}

		out = append(out, Article{Title: title, URL: href})
	}
	return out
}
