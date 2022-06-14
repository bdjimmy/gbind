package gbind

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// StructValidator StructValidator
type StructValidator interface {
	ValidateStruct(interface{}) error
}

// Validator Validator
var Validator StructValidator = &defaultValidator{}

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("validate")
	})
}

type sliceValidateError []error

func (err sliceValidateError) Error() string {
	var errMsgs []string
	for i, e := range err {
		if e == nil {
			continue
		}
		errMsgs = append(errMsgs, fmt.Sprintf("[%d]: %s", i, e.Error()))
	}
	return strings.Join(errMsgs, "\n")
}

var _ StructValidator = &defaultValidator{}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return v.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return v.validateStruct(obj)
	// case reflect.Slice, reflect.Array:
	//	count := value.Len()
	//	validateRet := make(sliceValidateError, 0)
	//	for i := 0; i < count; i++ {
	//		if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
	//			validateRet = append(validateRet, err)
	//		}
	//	}
	//	if len(validateRet) == 0 {
	//		return nil
	//	}
	//	return validateRet
	default:
		return nil
	}
}

// ValidateStruct receives struct type
func (v *defaultValidator) validateStruct(obj interface{}) error {
	v.lazyinit()
	// v.validate.RegisterValidation()
	return v.validate.Struct(obj)
}

func (v *defaultValidator) registerCustomValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	v.lazyinit()
	return v.validate.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}
