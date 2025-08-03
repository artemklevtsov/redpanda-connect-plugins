package schedule

import (
	"github.com/redpanda-data/benthos/v4/public/service"
)

func inputConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Utility").
		Summary("Reads messages from a child input on a schedule").
		Description("").
		Fields(
			service.NewInputField("input").
				Description("The child input to consume from."),
			service.NewDurationField("interval").
				Description("Interval at which the child input should be consumed."),
		)
}
