package logs

import "github.com/redpanda-data/benthos/v4/public/service"

func inputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("api", "http", "yandex").
		Summary("Creates an input that fetch Yandex.Metrika API logs data.").
		Fields(
			service.NewIntField("counter_id").
				Description("Yandex.Metrika Counter ID").
				Example(44147844),
			service.NewStringEnumField("source", "visits", "hits").
				Description("Log source.").
				Examples("visits", "hits"),
			service.NewStringListField("fields").
				Description("A list of fields.").
				Example([]string{"ym:s:dateTime", "ym:s:visitID", "ym:s:pageViews", "ym:s:isNewUser", "ym:s:counterUserIDHash"}),
			service.NewStringField("token").
				Description("Yandex.Metrika API token").
				Secret(),
			service.NewStringField("date1").
				Description("Start date of the sample period in YYYY-MM-DD format.").
				Default("6daysAgo").
				Optional(),
			service.NewStringField("date2").
				Description("End date of the sample period in YYYY-MM-DD format.").
				Default("today").
				Optional(),
			service.NewBoolField("format_keys").
				Description("Format output JSON keys.").
				Optional(),
		)
}
