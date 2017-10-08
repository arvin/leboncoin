package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"
)

func server(dbURL string) http.Handler {
	db := sqlx.MustConnect("postgres", dbURL)
	t, err := loadTemplates()
	if err != nil {
		log.Fatal(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var departments, cities, districts []string
		if err := db.Select(&departments, queries["select_departments.sql"]); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.Select(&cities, queries["select_cities.sql"]); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.Select(&districts, queries["select_districts.sql"]); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		announces, err := selectAnnounces(db, q)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := t.ExecuteTemplate(w, "index.html", struct {
			Announces                      []Announce
			Departments, Cities, Districts []string
			Query                          url.Values
		}{Query: q, Departments: departments, Cities: cities, Districts: districts, Announces: announces}); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func selectAnnounces(db *sqlx.DB, q url.Values) ([]Announce, error) {
	departments := q["departments"]
	cities := q["cities"]
	districts := q["districts"]

	var (
		query string
		args  []interface{}
		err   error
	)

	// seven cases
	// 1. departments + cities + districts
	// 2. departments + cities
	// 3. cities + districts
	// 4. departments
	// 5. cities
	// 6. districts
	// 7. all
	switch {
	case departments != nil && cities != nil && districts != nil:
		query, args, err = sqlx.In(queries["select_where_department_or_city_or_district.sql"], departments, cities, districts)
	case departments != nil && cities != nil:
		query, args, err = sqlx.In(queries["select_where_department_or_city.sql"], departments, cities)
	case cities != nil && districts != nil:
		query, args, err = sqlx.In(queries["select_where_city_or_district.sql"], cities, districts)
	case departments != nil:
		query, args, err = sqlx.In(queries["select_where_department.sql"], departments)
	case cities != nil:
		query, args, err = sqlx.In(queries["select_where_city.sql"], cities)
	case districts != nil:
		query, args, err = sqlx.In(queries["select_where_district.sql"], districts)
	default:
		query = queries["select.sql"]
	}
	if err != nil {
		return nil, err
	}

	var announces []Announce
	if err := db.Select(&announces, db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return announces, nil

}

func loadTemplates() (*template.Template, error) {
	t := template.New("").Funcs(
		template.FuncMap{"contains": contains},
	)
	for k, v := range templates {
		_, err := t.New(k).Parse(v)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if str == s {
			return true
		}
	}
	return false
}
