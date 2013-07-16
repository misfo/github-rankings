package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
)

const (
	oneDay   = 24 * 60 * 60 // in seconds
	hostName = "https://github.com"
)

var listTemp = template.Must(template.ParseFiles("list.html"))

type language struct {
	Name string
	Url  string
	rank int
}

func languagePages() []language {
	doc, e := goquery.NewDocument(hostName + "/languages")
	if e != nil {
		panic(e.Error())
	}
	languageSel := doc.Find(".all_languages li a")

	pages := make([]language, languageSel.Size())
	languageSel.Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		pages[i] = language{Name: s.Text(), Url: hostName + href}
	})
	return pages
}

func rank(languageUrl string) int {
	doc, e := goquery.NewDocument(languageUrl)
	if e != nil {
		panic(e.Error())
	}

	h1 := doc.Find(".pagehead h1").Text()
	re := regexp.MustCompile("the( #(\\d+))? most popular language")
	match := re.FindStringSubmatch(h1)
	if len(match) != 3 {
		panic("Unexpected page header: " + h1)
	}

	if match[2] == "" {
		return 1
	} else {
		rank, err := strconv.ParseInt(match[2], 10, 0)
		if err != nil {
			panic(err.Error())
		}
		return int(rank)
	}
}

func languageRankings() []language {
	pages := languagePages()
	numPages := len(pages)
	languageChan := make(chan language)
	for _, page := range pages {
		go func(page language) {
			page.rank = rank(page.Url)
			languageChan <- page
		}(page)
	}

	languages := make([]language, numPages)
	for i := 0; i < numPages; i++ {
		lang := <-languageChan
		if lang.rank > numPages {
			extendedLanguages := make([]language, lang.rank)
			copy(extendedLanguages, languages)
			languages = extendedLanguages
		}
		languages[lang.rank-1] = lang
	}
	return languages
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", oneDay))
	languages := languageRankings()
	listTemp.Execute(w, languages)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
