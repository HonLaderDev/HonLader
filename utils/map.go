package utils

import (
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

// MarshalMap 使用 mapstructure tag 将结构体转换为 map。
func MarshalMap(value any) (map[string]any, error) {
	encoded := encodeMapValue(reflect.ValueOf(value))
	result, ok := encoded.(map[string]any)
	if !ok {
		return map[string]any{}, nil
	}
	return result, nil
}

// UnmarshalMap 使用 mapstructure tag 将 map 写入结构体。
func UnmarshalMap(data map[string]any, value any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       stringToTimeHookFunc(),
		Metadata:         nil,
		Result:           value,
		Squash:           true,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		MatchName:        matchMapstructureName,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(data)
}

func stringToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String || to != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		text := data.(string)
		if text == "" {
			return time.Time{}, nil
		}
		return time.Parse(time.RFC3339Nano, text)
	}
}

func matchMapstructureName(mapKey, fieldName string) bool {
	return normalizeMapName(mapKey) == normalizeMapName(fieldName)
}

func normalizeMapName(name string) string {
	name = strings.ReplaceAll(name, "_", "")
	return strings.ToLower(name)
}

func encodeMapValue(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}
	for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	if value.Type() == reflect.TypeOf(time.Time{}) {
		t := value.Interface().(time.Time)
		if t.IsZero() {
			return ""
		}
		return t.Format(time.RFC3339Nano)
	}

	switch value.Kind() {
	case reflect.Struct:
		return encodeStruct(value)
	case reflect.Slice, reflect.Array:
		result := make([]any, 0, value.Len())
		for i := range value.Len() {
			result = append(result, encodeMapValue(value.Index(i)))
		}
		return result
	case reflect.Map:
		result := make(map[string]any, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				continue
			}
			result[key.String()] = encodeMapValue(iter.Value())
		}
		return result
	default:
		return value.Interface()
	}
}

func encodeStruct(value reflect.Value) map[string]any {
	valueType := value.Type()
	result := make(map[string]any, value.NumField())
	for i := range value.NumField() {
		field := valueType.Field(i)
		if !field.IsExported() {
			continue
		}
		name, squash, skip := mapstructureFieldName(field)
		if skip {
			continue
		}
		encoded := encodeMapValue(value.Field(i))
		if squash {
			if nested, ok := encoded.(map[string]any); ok {
				for key, nestedValue := range nested {
					result[key] = nestedValue
				}
			}
			continue
		}
		result[name] = encoded
	}
	return result
}

func mapstructureFieldName(field reflect.StructField) (name string, squash bool, skip bool) {
	tag := field.Tag.Get("mapstructure")
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" {
		name = field.Name
	}
	for _, part := range parts[1:] {
		if part == "squash" {
			squash = true
		}
	}
	return name, squash, false
}
