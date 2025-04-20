package bloblang

import (
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

func init() {
	method := "uuid_v5"

	spec := bloblang.NewPluginSpec().
		Description(`Retruns UUID version 5 for the given string. Available predefined namespaces: `+"`dns`, `url`, `oid`, `x500`"+`.`).
		Example("", `root.id = "example".uuid_v5()`, [2]string{`example`, `{"id": "feb54431-301b-52bb-a6dd-e1e93e81bb9e"}`}).
		Param(bloblang.NewStringParam("ns").Optional())

	ctor := func(args *bloblang.ParsedParams) (bloblang.Method, error) {
		ns, err := args.GetOptionalString("ns")
		if err != nil {
			return nil, err
		}

		return func(v any) (any, error) {
			if v == nil {
				return nil, nil
			}

			var uns uuid.UUID
			if ns == nil {
				uns = uuid.Nil
			} else {
				switch *ns {
				case "dns", "DNS":
					uns = uuid.NamespaceDNS
				case "url", "URL":
					uns = uuid.NamespaceURL
				case "oid", "OID":
					uns = uuid.NamespaceOID
				case "x500", "X500":
					uns = uuid.NamespaceX500
				default:
					uns, err = uuid.FromString(*ns)
					if err != nil {
						return nil, fmt.Errorf("invalid ns param for the %s: %q", method, *ns)
					}
				}
			}

			switch t := v.(type) {
			case string:
				return uuid.NewV5(uns, t).String(), nil
			case []byte:
				return uuid.NewV5(uns, string(t)).String(), nil
			}

			return nil, fmt.Errorf("invalid input type for the %s: %T", method, v)
		}, nil
	}

	err := bloblang.RegisterMethodV2(method, spec, ctor)
	if err != nil {
		panic(err)
	}
}
