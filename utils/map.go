package utils

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

// MarshalMap 使用 mapstructure tag 将结构体转换为 map。
func MarshalMap(value any) (map[string]any, error) {
	result := make(map[string]any)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       mapEncodeHook(),
		Metadata:         nil,
		Result:           &result,
		Squash:           true,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(value); err != nil {
		return nil, err
	}
	return result, nil
}

// UnmarshalMap 使用 mapstructure tag 将 map 写入结构体。
func UnmarshalMap(data map[string]any, value any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
		Metadata:         nil,
		Result:           value,
		Squash:           true,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(data)
}

func mapEncodeHook() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from == reflect.TypeOf(time.Time{}) && to.Kind() == reflect.Interface {
			t := data.(time.Time)
			if t.IsZero() {
				return "", nil
			}
			return t.Format(time.RFC3339Nano), nil
		}
		return data, nil
	}
}
