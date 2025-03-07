package utils

import "github.com/go-viper/mapstructure/v2"

func StructToMap(input any) (map[string]any, error) {
	output := make(map[string]any)

	conf := &mapstructure.DecoderConfig{
		Result:  &output,
		TagName: "json",
		Squash:  true,
	}

	dec, err := mapstructure.NewDecoder(conf)
	if err != nil {
		return nil, err
	}

	if err := dec.Decode(input); err != nil {
		return nil, err
	}

	return output, nil
}
