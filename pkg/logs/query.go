package logs

import (
	"strings"
)

type apiQuery struct {
	Source      string   `json:"source"`
	Date1       string   `json:"date1"`
	Date2       string   `json:"date2"`
	Fields      []string `json:"fields"`
	Attribution string   `json:"attribution"`
}

func (q *apiQuery) Params() map[string]string {
	m := make(map[string]string)
	m["source"] = q.Source
	m["date1"] = q.Date1
	m["date2"] = q.Date2
	m["fields"] = strings.Join(q.Fields, ",")
	m["attribution"] = q.Attribution

	for k := range m {
		if len(m[k]) == 0 {
			delete(m, k)
		}
	}

	return m
}
