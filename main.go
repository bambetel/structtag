package main

// Based on
// https://medium.com/swlh/custom-struct-field-tags-and-validation-in-golang-9a7aeedcdc5b

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

const tagValidate = "validate"

// TODO: validation by type or meaning of field:
//
//	email
//	age
type User struct {
	Id       int    `json:"id"`
	Name     string `validate:"min=5,max=32" json:"name"`
	Email    string `validate:"regexemail" json:"email"`
	Age      int    `validate:"min=18"`
	Password string `validate:"length=24" json:"-"`
}

func main() {
	u := User{
		Id:       123,
		Name:     "Testowe",
		Email:    "test@example.com",
		Password: "123456123456123456123456",
	}
	r := reflect.TypeOf(u)

	fmt.Printf("Name: %v, kind: %v\n", r.Name(), r.Kind())

	err := json.NewEncoder(os.Stdout).Encode(u)
	if err != nil {
		panic("Struct encoding error.")
	}

	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		tag := f.Tag.Get(tagValidate)
		fmt.Printf("%v [%v] %q %v\n", f.Name, f.Type.Name(), tag, f.Type)
	}

	ValidateStruct(u)
}

type Validator interface {
	Validate(interface{}) (bool, error)
}

func ValidateStruct(s interface{}) []error {
	res := []error{}
	v := reflect.ValueOf(s)

	for i, r := 0, reflect.TypeOf(s); i < r.NumField(); i++ {
		f := r.Field(i)
		tag := f.Tag.Get(tagValidate)
		if tr := strings.TrimSpace(tag); tr == "" || tr == "-" {
			continue
		}

		validator, err := getValidatorFor(f.Type.Name())
		fmt.Printf("Get validator result: %v, %v\n", validator, err)
		if validator == nil {
			continue
		}
		valid, err := validator.Validate(v.Field(i).Interface())
		fmt.Printf("Validation result: %v %v\n", valid, err)

		if !valid && err != nil {
			res = append(res, fmt.Errorf("Struct field %q validation error: %s", v.Type().Field(i).Name, err.Error()))
		}

	}
	fmt.Printf("%v\n", v)
	return res
}

// getting validator by type name
// TODO: smarter
func getValidatorFor(t string) (Validator, error) {
	fmt.Printf("Get validator for: %q\n", t)
	switch t {
	case "string":
		// hardcoded values!
		return StringValidator{5, 12, true}, nil
	}

	return nil, fmt.Errorf("Validator for type %v not found", t)
}

// StringValidator validates string presence and/or its length.
type StringValidator struct {
	Min      int
	Max      int
	Required bool
}

func (v StringValidator) Validate(val interface{}) (bool, error) {
	l := len(val.(string))
	if l == 0 && v.Required {
		return false, fmt.Errorf("Field required")
	}
	if l < v.Min {
		return false, fmt.Errorf("should be at least %v chars long", v.Min)
	}
	if v.Max >= v.Min && l > v.Max {
		return false, fmt.Errorf("should be less than %v chars long", v.Max)
	}
	return true, nil
}
