package goals

import "github.com/redpanda-data/benthos/v4/public/service"

func inputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("api", "http", "yandex").
		Summary("Creates an input that fetch Yandex.Metrika API goals.").
		Fields(
			service.NewIntField("counter_id").
				Description("Yandex.Metrika Counter ID").
				Example(44147844),
			service.NewStringField("token").
				Description("Yandex.Metrika API token").
				Secret().
				Optional(),
		)
}
