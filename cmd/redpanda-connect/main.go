package main

import (
	"context"

	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/redpanda-data/connect/v4/public/license"

	// _ "github.com/redpanda-data/connect/public/bundle/free/v4"

	// Import full suite of FOSS and Enterprise connect components
	_ "github.com/redpanda-data/connect/public/bundle/enterprise/v4"

	// _ "github.com/redpanda-data/connect/v4/public/components/io"
	// _ "github.com/redpanda-data/connect/v4/public/components/pure"
	// _ "github.com/redpanda-data/connect/v4/public/components/pure/extended"
	// _ "github.com/redpanda-data/connect/v4/public/components/sql"

	// _ "github.com/artemklevtsov/redpanda-connect-plugins/pkg/bloblang"
	_ "github.com/artemklevtsov/redpanda-connect-plugins/pkg/input/yandex/appmetrika"
	_ "github.com/artemklevtsov/redpanda-connect-plugins/pkg/input/yandex/metrika"
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
		service.CLIOptOnConfigParse(func(conf *service.ParsedConfig) error {
			license.LocateLicense(conf.Resources())
			return nil
		}),
		service.CLIOptSetVersion(Version, DateBuilt),
		service.CLIOptSetDefaultConfigPaths(
			"benthos.yaml",
			"connect.yaml",
			"redpanda-connect.yaml",
		),
	)
}
