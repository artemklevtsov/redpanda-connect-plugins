package stat_table

import "github.com/redpanda-data/benthos/v4/public/service"

func inputConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("api", "http", "yandex").
		Summary("Creates an input that fetch Yandex.Metrika API report data.").
		Fields(
			service.NewStringField("token").
				Description("Yandex.AppMetrika API token").
				Secret().
				Optional(),
			service.NewIntListField("ids").
				Description("Yandex.AppMetrika Counter IDs").
				Example([]int{44147844, 2215573}),
			service.NewStringListField("metrics").
				Description("A list of metrics.").
				Example([]string{"ym:s:pageviews", "ym:s:visits", "ym:s:users"}),
			service.NewStringListField("dimensions").
				Description("A list of dimensions.").
				Optional(),
			service.NewStringField("date1").
				Description("Start date of the sample period in YYYY-MM-DD format.").
				Default("6daysAgo").
				Optional(),
			service.NewStringField("date2").
				Description("End date of the sample period in YYYY-MM-DD format.").
				Default("today").
				Optional(),
			service.NewStringField("filters").
				Description("Segmentation filter.").
				Optional(),
			service.NewStringListField("sort").
				Description("A list of dimensions and metrics to use for sorting.").
				Optional(),
			service.NewStringField("accuracy").
				Description("Sample size for the report.").
				Optional(),
			service.NewStringField("lang").
				Description("Language.").
				Example("en").
				Optional(),
			service.NewStringField("preset").
				Description("Report preset.").
				Example("sources_summary").
				Optional(),
			service.NewStringField("timezone").
				Description("Time zone in ±hh:mm format within the range of [-23:59; +23:59]").
				Example("+03:00").
				Optional(),
			service.NewStringListField("direct_client_logins").
				Description("A list of usernames of Yandex Direct clients").
				Optional(),
		)
}
