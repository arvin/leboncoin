package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/html/charset"
)

func fetch(dbURL string) {
	announces, err := fetchAnnounces()
	if err != nil {
		log.Fatal(err)
	}

	db := sqlx.MustConnect("postgres", dbURL)
	for _, a := range announces {
		db.MustExec(queries["insert.sql"],
			a.Href,
			a.Title,
			a.Price,
			a.PublishedAt,
			a.Urgent,
			a.Department,
			a.City,
			a.District)
	}
}

func fetchAnnounces() ([]Announce, error) {
	// f, err := os.Open("ile_de_france.html")
	// if err != nil {
	// 	return nil, err
	// }
	// defer f.Close()
	// contentType := "text/html; charset=windows-1252"
	// reader := f

	resp, err := http.Get("https://www.leboncoin.fr/colocations/offres/ile_de_france/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	reader := resp.Body

	r, err := charset.NewReader(reader, contentType)
	if err != nil {
		return nil, err
	}

	return newAnnounces(r)
}

func newAnnounces(r io.Reader) ([]Announce, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var (
		announces []Announce
		parseErr  error
	)
	doc.Find("#listingAds > section > section > ul > li").EachWithBreak(func(i int, s *goquery.Selection) bool {
		var a Announce

		a.Href, _ = s.Find("a").Attr("href")
		a.Title = strings.TrimSpace(s.Find("a > section > h2").Text())

		a.Department, a.City, a.District, err = parseLocation(strings.Join(strings.Fields(s.Find("a > section > p:nth-child(3)").Text()), " "))
		if err != nil {
			parseErr = err
			return false
		}

		a.Price, err = parsePrice(strings.Join(strings.Fields(s.Find("a > section > h3").Text()), " "))
		if err != nil {
			parseErr = err
			return false
		}

		publishedAt := strings.TrimSpace(s.Find("a > section > aside > p").Text())
		if strings.HasPrefix(publishedAt, "Urgent") {
			a.Urgent = true
			publishedAt = strings.TrimSpace(strings.TrimPrefix(publishedAt, "Urgent"))
		}
		a.PublishedAt, err = parseTime(publishedAt)
		if err != nil {
			parseErr = err
			return false
		}

		announces = append(announces, a)
		return true
	})

	// TODO: return all parse errors, not just the first one
	return announces, parseErr
}

// TODO: remove global var and init() func?
var paris *time.Location

func init() {
	var err error
	paris, err = time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}
}

func parseTime(s string) (time.Time, error) {
	var year, day, hour, minute int
	var month time.Month

	if strings.HasPrefix(s, "Aujourd'hui, ") {
		year, month, day = time.Now().In(paris).Date()
		s = strings.TrimPrefix(s, "Aujourd'hui, ")

		split := strings.Split(s, ":")
		if len(split) != 2 {
			return time.Time{}, fmt.Errorf("couldn't parse clock: %v", s)
		}

		var err error
		hour, err = strconv.Atoi(split[0])
		if err != nil {
			return time.Time{}, err
		}

		minute, err = strconv.Atoi(split[1])
		if err != nil {
			return time.Time{}, err
		}
	}

	// TODO: handle case where the prefix is not "Aujourd'hui"
	return time.Date(year, month, day, hour, minute, 0, 0, paris), nil
}

func parsePrice(s string) (*int, error) {
	join := strings.Join(strings.Fields(s), "")
	if join == "" {
		return nil, nil
	}

	for i, c := range join {
		if !unicode.IsDigit(c) {
			join = join[:i]
			break
		}
	}

	price, err := strconv.Atoi(join)
	if err != nil {
		return nil, err
	}

	return &price, nil
}

func parseLocation(s string) (string, string, string, error) {
	var department, city, district string
	if split := strings.Split(s, "/"); len(split) == 2 {
		city = strings.TrimSpace(split[0])
		department = strings.TrimSpace(split[1])
		return department, city, district, nil
	}

	fields := strings.Fields(s)
	switch len(fields) {
	case 2:
		city = fields[0]
		district = fields[1]
		return department, city, district, nil
	case 1:
		department = fields[0]
		return department, city, district, nil
	default:
		return "", "", "", fmt.Errorf("couldn't parse location: %v", s)
	}
}
