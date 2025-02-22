package misc

import (
	"regexp"

	"github.com/iancoleman/strcase"
)

func FormatKey(k string) string {
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
