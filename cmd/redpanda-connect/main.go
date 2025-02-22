package main

import (
	"context"

	"github.com/redpanda-data/benthos/v4/public/service"

	_ "github.com/redpanda-data/connect/public/bundle/free/v4"

	// _ "github.com/redpanda-data/connect/v4/public/components/io"
	// _ "github.com/redpanda-data/connect/v4/public/components/pure"
	// _ "github.com/redpanda-data/connect/v4/public/components/pure/extended"
	// _ "github.com/redpanda-data/connect/v4/public/components/sql"

	_ "github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/goals"
	_ "github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/logs"
	_ "github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/stat_table"
)

var (
	// Version version set at compile time.
	Version string
	// DateBuilt date built set at compile time.
	DateBuilt string
)

func main() {
	service.RunCLI(
		context.Background(),
		service.CLIOptSetVersion(Version, DateBuilt),
		service.CLIOptSetDefaultConfigPaths(
			"benthos.yaml",
			"connect.yaml",
			"redpanda-connect.yaml",
		),
	)
}
