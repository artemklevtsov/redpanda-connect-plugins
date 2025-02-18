package stat_table

import (
	"regexp"

	"github.com/iancoleman/strcase"
)

func formatKey(k string) string {
	k = regexp.MustCompile("^ym:.*:").
		ReplaceAllString(k, "")

	// snake case exclusions
	switch k {
	case "watchIDs":
		return "watch_ids"
	case "iFrame":
		return "iframe"
	}

	k = strcase.ToSnake(k)

	return k
}
