package gemini

import (
	"encoding/json"
	"fmt"
	"reflect"

	"google.golang.org/genai"
)

func DecodeJsonContent[T any](data json.RawMessage) ([]T, error) {
	var output []T
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return output, nil
}

func GenerateSchema(v interface{}) map[string]*genai.Schema {
	schema := make(map[string]*genai.Schema)
	t := reflect.TypeOf(v)

	// ポインタの場合はデリファレンス
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 構造体でなければエラー
	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// 型を genai のスキーマタイプに変換
		schema[jsonTag] = &genai.Schema{
			Type: goTypeToGenaiType(field.Type),
		}
	}

	return schema
}

// goTypeToGenaiType: Go の型を genai の型に変換
func goTypeToGenaiType(t reflect.Type) genai.Type {
	switch t.Kind() {
	case reflect.String:
		return genai.TypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return genai.TypeInteger
	case reflect.Float32, reflect.Float64:
		return genai.TypeNumber
	case reflect.Bool:
		return genai.TypeBoolean
	case reflect.Slice, reflect.Array:
		return genai.TypeArray
	case reflect.Struct:
		return genai.TypeObject
	default:
		return genai.TypeUnspecified
	}
}
