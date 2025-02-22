package stat_table

import (
	"fmt"
	"strconv"
	"strings"
)

type apiQuery struct {
	IDs          []int    `json:"ids"`
	Date1        string   `json:"date1,omitempty"`
	Date2        string   `json:"date2,omitempty"`
	Dimensions   []string `json:"dimensions,omitempty"`
	Metrics      []string `json:"metrics"`
	Sort         []string `json:"sort,omitempty"`
	Limit        int      `json:"limit"`
	Offset       int      `json:"offset"`
	Filters      string   `json:"filters,omitempty"`
	Lang         string   `json:"lang,omitempty"`
	Preset       string   `json:"preset,omitempty"`
	Timezone     string   `json:"timezone,omitempty"`
	DirectLogins []string `json:"direct_client_logins"`
}

func (q *apiQuery) Params() map[string]string {
	m := make(map[string]string)
	m["ids"] = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(q.IDs)), ","), "[]")
	m["filters"] = q.Filters
	m["date1"] = q.Date1
	m["date2"] = q.Date2
	m["dimensions"] = strings.Join(q.Dimensions, ",")
	m["metrics"] = strings.Join(q.Metrics, ",")
	m["lang"] = q.Lang
	m["preset"] = q.Preset
	m["sort"] = strings.Join(q.Sort, ",")
	m["timezone"] = q.Timezone
	m["limit"] = strconv.Itoa(q.Limit)
	m["offset"] = strconv.Itoa(q.Offset)

	for k := range m {
		if len(m[k]) == 0 {
			delete(m, k)
		}
	}

	return m
}
