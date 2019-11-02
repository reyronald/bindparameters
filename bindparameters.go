// Package bindparameters exposes a small utility function that will automatically bind URL parameters,
// query string parameters and/or body JSON payloads from a HTTP request into your own types without
// any need for manual decoding, marshalling or lookups, through a user-provided callback.
package bindparameters

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Into will automatically bind or map parameters from the HTTP request `r` into the arguments of `fn`.
// `fn` must be a function with either one or two arguments.
// The first argument should be a struct with fields that map
// to URL and query string parameters. The second argument is
// optional and will be used to bind/map the JSON payload of the request into it.
// If your endpoint doesn't have any URL or query string parameters
// (or you don't need to access them in your handler), you still need
// to provide the first argument to the function, but in that case you can pass `nil`.
// See the README for examples.
func Into(
	r *http.Request,
	getURLParam func(key string) string,
	fn interface{},
) {
	// Input validation
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("expects a function")
	}

	inputs := getInputs(fnType)
	if inputLen := len(inputs); inputLen != 1 && inputLen != 2 {
		panic("there should be only one or two arguments")
	}

	input := inputs[0]
	if input.Kind() != reflect.Struct {
		panic("argument must be a struct")
	}

	fields := getFields(input)
	fieldTypes := getStructFieldsTypes(fields)
	foundComplexTypes := filterComplexTypes(fieldTypes)
	if len(foundComplexTypes) > 0 {
		panic("there cannot be any complex types in the first argument's struct")
	}

	// Work //
	inputValue := reflect.New(input).Elem()

	// URLParams
	for _, field := range fields {
		urlParam := getURLParam(field.Name)
		convertToKindAndSetValueIn(urlParam, field.Type.Kind(), inputValue.FieldByName(field.Name))
	}

	// Query string
	for _, field := range fields {
		var foundValue []string
		for k, value := range r.URL.Query() {
			normalizedKey := strings.TrimSuffix(
				strings.ToLower(k),
				"[]",
			)

			if normalizedKey == strings.ToLower(field.Name) {
				foundValue = value
				break
			}
		}

		if len(foundValue) > 0 && field.Type.Kind() != reflect.Slice {
			queryParam := foundValue[0]
			fieldTypeKind := field.Type.Kind()
			fieldValue := inputValue.FieldByName(field.Name)
			convertToKindAndSetValueIn(queryParam, fieldTypeKind, fieldValue)
		} else if field.Type.Kind() == reflect.Slice {
			lenValue := len(foundValue)
			fieldTypeKind := field.Type.Elem().Kind()
			s := reflect.MakeSlice(field.Type, lenValue, lenValue)
			for i := 0; i < lenValue; i++ {
				convertToKindAndSetValueIn(
					foundValue[i],
					fieldTypeKind,
					s.Index(i),
				)
			}

			inputValue.FieldByName(field.Name).Set(s)
		}
	}

	// Request body
	var complexTypeValue interface{}
	hasBody := len(inputs) == 2
	if hasBody {
		complexType := inputs[1]
		complexTypeValue = reflect.New(complexType).Interface()
		err := json.NewDecoder(r.Body).Decode(&complexTypeValue)
		if err != nil {
			panic(err)
		}
	}

	// Call fn
	if fnValue := reflect.ValueOf(fn); hasBody {
		fnValue.Call([]reflect.Value{
			inputValue,
			reflect.Indirect(reflect.ValueOf(complexTypeValue)),
		})
	} else {
		fnValue.Call([]reflect.Value{
			inputValue,
		})
	}
}

func convertToKindAndSetValueIn(valueToSet string, kind reflect.Kind, dstValue reflect.Value) {
	if valueToSet != "" {
		switch kind {
		case reflect.Bool:
			b, _ := strconv.ParseBool(valueToSet)
			dstValue.SetBool(b)
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			fallthrough
		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			i, _ := strconv.Atoi(valueToSet)
			dstValue.SetInt(int64(i))
		case reflect.Float32:
			f, _ := strconv.ParseFloat(valueToSet, 32)
			dstValue.SetFloat(f)
		case reflect.Float64:
			f, _ := strconv.ParseFloat(valueToSet, 64)
			dstValue.SetFloat(f)
		case reflect.String:
			dstValue.SetString(valueToSet)
		default:
			panic("unsupported field kind " + kind.String())
		}
	}
}

func getFields(input reflect.Type) []reflect.StructField {
	fields := []reflect.StructField{}
	for i := 0; i < input.NumField(); i++ {
		fields = append(fields, input.Field(i))
	}
	return fields
}

func filter(vs []reflect.Type, f func(reflect.Type) bool) []reflect.Type {
	vsf := []reflect.Type{}
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func index(vs []reflect.Kind, t reflect.Kind) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func include(vs []reflect.Kind, t reflect.Kind) bool {
	return index(vs, t) >= 0
}

func getStructFieldsTypes(fields []reflect.StructField) []reflect.Type {
	s := []reflect.Type{}
	for _, f := range fields {
		s = append(s, f.Type)
	}
	return s
}

func filterComplexTypes(inputs []reflect.Type) []reflect.Type {
	supportedTypes := []reflect.Kind{
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.Array,
		reflect.Slice,
		reflect.String,
	}
	foundComplexTypes := filter(inputs, func(input reflect.Type) bool {
		isSuppportedType := include(supportedTypes, input.Kind())
		return !isSuppportedType
	})
	return foundComplexTypes
}

func getInputs(fnType reflect.Type) []reflect.Type {
	if fnType.Kind() != reflect.Func {
		panic("expects a function")
	}

	inputs := []reflect.Type{}
	for i := 0; i < fnType.NumIn(); i++ {
		inputs = append(inputs, fnType.In(i))
	}
	return inputs
}
