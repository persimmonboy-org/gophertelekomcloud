package extract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

func intoPtr(body, to interface{}, label string) error {
	if label == "" {
		return Into(body, &to)
	}

	var m map[string]interface{}
	err := Into(body, &m)
	if err != nil {
		return err
	}

	b, err := JsonMarshal(m[label])
	if err != nil {
		return err
	}

	toValue := reflect.ValueOf(to)
	if toValue.Kind() == reflect.Ptr {
		toValue = toValue.Elem()
	}

	switch toValue.Kind() {
	case reflect.Slice:
		typeOfV := toValue.Type().Elem()
		if typeOfV.Kind() == reflect.Struct {
			if typeOfV.NumField() > 0 && typeOfV.Field(0).Anonymous {
				newSlice := reflect.MakeSlice(reflect.SliceOf(typeOfV), 0, 0)

				for _, v := range m[label].([]interface{}) {
					// For each iteration of the slice, we create a new struct.
					// This is to work around a bug where elements of a slice
					// are reused and not overwritten when the same copy of the
					// struct is used:
					//
					// https://github.com/golang/go/issues/21092
					// https://github.com/golang/go/issues/24155
					// https://play.golang.org/p/NHo3ywlPZli
					newType := reflect.New(typeOfV).Elem()

					b, err := JsonMarshal(v)
					if err != nil {
						return err
					}

					// This is needed for structs with an UnmarshalJSON method.
					// Technically this is just unmarshalling the response into
					// a struct that is never used, but it's good enough to
					// trigger the UnmarshalJSON method.
					for i := 0; i < newType.NumField(); i++ {
						s := newType.Field(i).Addr().Interface()

						// Unmarshal is used rather than NewDecoder to also work
						// around the above-mentioned bug.
						err = json.Unmarshal(b, s)
						if err != nil {
							continue
						}
					}

					newSlice = reflect.Append(newSlice, newType)
				}

				// "to" should now be properly modeled to receive the
				// JSON response body and unmarshal into all the correct
				// fields of the struct or composed extension struct
				// at the end of this method.
				toValue.Set(newSlice)
			}
		}
	case reflect.Struct:
		typeOfV := toValue.Type()
		if typeOfV.NumField() > 0 && typeOfV.Field(0).Anonymous {
			for i := 0; i < toValue.NumField(); i++ {
				toField := toValue.Field(i)
				if toField.Kind() == reflect.Struct {
					s := toField.Addr().Interface()
					err = json.NewDecoder(bytes.NewReader(b)).Decode(s)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	err = json.Unmarshal(b, &to)
	return err
}

func JsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(false)
	err := enc.Encode(t)
	return buffer.Bytes(), err
}

func Into(body interface{}, to interface{}) error {
	if reader, ok := body.(io.Reader); ok {
		if readCloser, ok := reader.(io.Closer); ok {
			defer readCloser.Close()
		}
		return json.NewDecoder(reader).Decode(to)
	}

	b, err := JsonMarshal(body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, to)

	return err
}

// IntoStructPtr will unmarshal the given body into the provided
// interface{} (to).
func IntoStructPtr(body, to interface{}, label string) error {
	t := reflect.TypeOf(to)
	if k := t.Kind(); k != reflect.Ptr {
		return fmt.Errorf("expected pointer, got %v", k)
	}
	switch t.Elem().Kind() {
	case reflect.Struct:
		return intoPtr(body, to, label)
	default:
		return fmt.Errorf("expected pointer to struct, got: %v", t)
	}
}

// IntoSlicePtr will unmarshal the provided body into the provided
// interface{} (to).
func IntoSlicePtr(body, to interface{}, label string) error {
	t := reflect.TypeOf(to)
	if k := t.Kind(); k != reflect.Ptr {
		return fmt.Errorf("expected pointer, got %v", k)
	}
	switch t.Elem().Kind() {
	case reflect.Slice:
		return intoPtr(body, to, label)
	default:
		return fmt.Errorf("expected pointer to slice, got: %v", t)
	}
}