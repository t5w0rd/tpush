package internal

import "github.com/mitchellh/mapstructure"

func Map2Struct(data interface{}, output interface{}) error {
	return mapstructure.Decode(data, output)
}
