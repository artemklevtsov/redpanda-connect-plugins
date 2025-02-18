package stat_table

import "github.com/redpanda-data/benthos/v4/public/service"

func inputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("api", "http", "yandex").
		Summary("Creates an input that fetch Yandex.Metrika API report.").
		Fields(
			service.NewIntListField("ids").
				Description("Yandex.Metrika Counter ID").
				Example([]int{44147844, 2215573}),
			service.NewStringListField("metrics").
				Description("A list of metrics.").
				Example([]string{"ym:s:pageviews", "ym:s:visits", "ym:s:users"}),
			service.NewStringField("token").
				Description("Yandex.Metrika API token").
				Secret().
				Optional(),
			service.NewStringListField("dimensions").
				Description("A list of dimensions.").
				Optional(),
			service.NewStringField("filters").
				Description("Segmentation filter.").
				Optional(),
			service.NewStringField("date1").
				Description("Start date of the sample period in YYYY-MM-DD format.").
				Default("6daysAgo").
				Optional(),
			service.NewStringField("date2").
				Description("End date of the sample period in YYYY-MM-DD format.").
				Default("today").
				Optional(),
			service.NewStringField("lang").
				Description("Language.").
				Optional(),
			service.NewStringField("preset").
				Description("Report preset.").
				Optional(),
			service.NewStringListField("sort").
				Description("A list of dimensions and metrics to use for sorting.").
				Optional(),
			service.NewStringField("timezone").
				Description("Time zone in ±hh:mm format within the range of [-23:59; +23:59]").
				Optional(),
			service.NewStringListField("direct_client_logins").
				Description("A list of usernames of Yandex Direct clients"),
			service.NewBoolField("format_keys").
				Description("Format output JSON keys.").
				Optional(),
		)
}
