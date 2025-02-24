package misc

import (
	"encoding/json"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ProcessValue(key, value string) any {
	var handler func(s string) any

	key = strings.TrimPrefix(key, "ym:s:")
	key = strings.TrimPrefix(key, "ym:pv:")

	switch key {
	case "watchIDs":
		handler = FixWatchIDs
	case "goalsDateTime":
		handler = FixArrayDateTime
	case
		// https://yandex.ru/dev/metrika/ru/logs/fields/hits
		"watchID", "pageViewID", "counterID", "clientID", "counterUserIDHash", "hasGCLID", "browserMajorVersion", "browserMinorVersion", "browserEngineVersion1", "browserEngineVersion2", "browserEngineVersion3", "browserEngineVersion4", "clientTimeZone", "cookieEnabled", "javascriptEnabled", "physicalScreenHeight", "physicalScreenWidth", "screenColors", "screenHeight", "screenOrientation", "screenWidth", "windowClientHeight", "windowClientWidth", "regionCityID", "regionCountryID", "isPageView", "isTurboPage", "isTurboApp", "iFrame", "link", "download", "notBounce", "artificial",
		// https://yandex.ru/dev/metrika/ru/logs/fields/visits
		"visitID", "isNewUser", "pageViews", "visitDuration", "bounce":
		handler = func(s string) any {
			return json.Number(s)
		}
	default:
		if len(value) >= 2 && value[0] == '[' && value[len(value)-1] == ']' {
			handler = ParseArray
		} else {
			handler = func(s string) any {
				return s
			}
		}
	}

	return handler(value)
}

func ParseArray(s string) any {
	if len(s) == 0 || s == `[]` {
		return s
	}

	if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
		var r []any
		if err := yaml.Unmarshal([]byte(s), &r); err != nil {
			return s
		}

		return r
	}

	return s
}

func FixArrayDateTime(s string) any {
	if len(s) == 0 || s == `[]` {
		return s
	}

	if len(s) > 2 && s[0] == '[' && s[len(s)-1] == ']' {
		var r []string
		if err := yaml.Unmarshal([]byte(s), &r); err != nil {
			return s
		}

		for i := range r {
			r[i] = strings.Trim(r[i], `\'`)
		}

		return r
	}

	return s
}

func FixWatchIDs(s string) any {
	if len(s) == 0 || s == `[]` {
		return s
	}

	if len(s) > 2 && s[0] == '[' && s[len(s)-1] == ']' {
		var r []json.Number
		if err := yaml.Unmarshal([]byte(s), &r); err != nil {
			return s
		}

		for i, v := range r {
			if len(v) > 0 && v[0] == '-' {
				n, err := v.Int64()
				if err != nil {
					return s
				}

				//nolint:gosec
				r[i] = json.Number(strconv.FormatUint(uint64(n), 10))
			}
		}

		return r
	}

	return s
}
