package gbind

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// uids string `gbind:"http.query.uid,default=1|2|3" err_msg:"uids'count should gte 1" validate:"gte=1"`
	defaultBindTag   = "gbind"
	defaultErrTag    = "err_msg"
	defaultSplitFlag = "|"
)

// Gbind contains the gbind settings and cache
type Gbind struct {
	// Bind tag name being used
	bindTagName string
	// Err tag name being used
	errTagName string
	// defaultSplitFlag for split default value
	defaultSplitFlag string
	// useNumberForJSON causes the Decoder to unmarshal a number into an interface{} as a
	// Number instead of as a float64.
	useNumberForJSON bool
	// local cache for *structType
	localCache *cache
	// tag execers facotry
	tagExcers *ExecerFactory
	// validator
	validator *defaultValidator
}

// Helper gbind so users can use the functions directly from the package
var defaultGbind = NewGbind()

// NewGbind returns a new instance of 'Gbind' with sane defaults.
func NewGbind() *Gbind {
	g := &Gbind{
		bindTagName:      defaultBindTag,
		errTagName:       defaultErrTag,
		defaultSplitFlag: defaultSplitFlag,
		useNumberForJSON: false,
		localCache:       newCache(),
		tagExcers:        NewExecerFactory(),
		validator:        &defaultValidator{},
	}
	g.tagExcers.Regitster("http", NewHTTPExecer)
	return g
}

// Bind parses the data interface and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Bind returns an Err.
func Bind(ctx context.Context, v interface{}, data interface{}) (context.Context, error) {
	return defaultGbind.bind(ctx, v, data, false)
}

// BindWithValidate parses the data interface and stores the result
// in the value pointed to by v and check whether the data meets the requirements.
// If v is nil or not a pointer, Bind returns an Err.
func BindWithValidate(ctx context.Context, v interface{}, data interface{}) (context.Context, error) {
	return defaultGbind.bind(ctx, v, data, true)
}

// SetBindTag allows you to change the tag name used in structs
func (g *Gbind) SetBindTag(tag string) {
	g.bindTagName = tag
}

// SetErrTag allows you to change the tag name used in structs
func (g *Gbind) SetErrTag(tag string) {
	g.errTagName = tag
}

// SetDefalutSplit allows you to change the splitFlag used in structs
func (g *Gbind) SetDefalutSplit(flag string) {
	g.defaultSplitFlag = flag
}

// RegisterBindFunc adds a bind Excer with the given name
func (g *Gbind) RegisterBindFunc(name string, fn NewExecer) {
	g.tagExcers.Regitster(name, fn)
}

// RegisterCustomValidation adds a validation with the given tag
//
// NOTES:
// - if the key already exists, the previous validation function will be replaced.
// - this method is not thread-safe it is intended that these all be registered prior to any validation
func (g *Gbind) RegisterCustomValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	return g.validator.registerCustomValidation(tag, fn, callValidationEvenIfNull...)
}

// UseNumberForJSON causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64.
func (g *Gbind) UseNumberForJSON(use bool) {
	g.useNumberForJSON = use
}

// Bind parses the data interface and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Bind returns an Err.
func (g *Gbind) Bind(ctx context.Context, v interface{}, data interface{}) (context.Context, error) {
	return g.bind(ctx, v, data, false)
}

// BindWithValidate parses the data interface and stores the result
// in the value pointed to by v and check whether the data meets the requirements.
// If v is nil or not a pointer, Bind returns an Err.
func (g *Gbind) BindWithValidate(ctx context.Context, v interface{}, data interface{}) (context.Context, error) {
	return g.bind(ctx, v, data, true)
}

func (g *Gbind) bind(ctx context.Context, v interface{}, data interface{}, validate bool) (context.Context, error) {
	rv := reflect.ValueOf(v)
	if err := g.checkValid(rv); err != nil {
		return ctx, err
	}
	st, err := g.compile(v)
	if err != nil {
		return ctx, err
	}
	// special case
	if st.hasJSONTag {
		st.parseJSON(data, v)
	}
	for _, f := range st.fields {
		if f.excer == nil {
			continue
		}
		ctx, err = f.excer.Exec(ctx, fieldByIndexs(rv, f.index), data, &f.defaultOpt)
		if err != nil {
			return ctx, err
		}
	}
	if validate {
		err = g.errMsg(st, g.validator.ValidateStruct(rv.Interface()))
	}
	return ctx, err
}

func (g *Gbind) compile(value interface{}) (*structType, error) {
	rt := reflect.TypeOf(value)
	if st, ok := g.localCache.get(rt); ok {
		return st, nil
	}
	st := &structType{
		g:          g,
		hasJSONTag: false,
		fields:     map[string]*fieldInfo{},
		errMap:     map[string]string{},
	}
	if err := st.deepTraverse(rt.Elem(), reflect.StructField{}, rt.Elem().Name(), []int{}); err != nil {
		return nil, err
	}
	st.g.localCache.set(rt, st)
	return st, nil
}

func (g *Gbind) checkValid(rv reflect.Value) error {
	if rv.Kind() != reflect.Ptr {
		s := "%q"
		if rv.Type() == nil {
			s = "%v"
		}
		return e("cannot bind to non-pointer "+s, rv.Type())
	}
	if rv.IsNil() {
		return e("cannot bind to a nil value of %q", rv.Type())
	}
	if rv.Elem().Type().Kind() != reflect.Struct {
		return e("binding must be a struct pointer")
	}
	return nil
}

func (g *Gbind) errMsg(st *structType, err error) error {
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			if msg, ok := st.errMap[e.Namespace()]; ok {
				return errors.New(msg)
			}
		}
	}
	return err
}

type structType struct {
	g          *Gbind
	hasJSONTag bool
	fields     map[string]*fieldInfo
	errMap     map[string]string
}

type fieldInfo struct {
	namespace   string
	index       []int
	structField reflect.StructField
	excer       Execer
	defaultOpt  DefaultOption
}

func fieldByIndexs(v reflect.Value, indexs []int) reflect.Value {
	for _, i := range indexs {
		v = reflect.Indirect(v).Field(i)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(deref(v.Type())))
			}
			v = reflect.Indirect(v)
		}
	}
	return v
}

func (sv *structType) getFieldByNS(v reflect.Value, ns string) reflect.Value {
	return fieldByIndexs(v, sv.fields[ns].index)
}

func (sv *structType) deepTraverse(rt reflect.Type, field reflect.StructField, ns string, index []int) error {
	if !field.Anonymous && field.PkgPath != "" {
		return nil
	}
	switch rt.Kind() {
	case reflect.Ptr:
		return sv.traversePtr(rt, field, ns, index)
	case reflect.Struct:
		return sv.traverseStruct(rt, field, ns, index)
	default:
		return sv.traverseField(rt, field, ns, index)
	}
}

func (sv *structType) traversePtr(rt reflect.Type, field reflect.StructField, ns string, index []int) error {
	return sv.deepTraverse(rt.Elem(), field, ns, index)
}

func (sv *structType) traverseStruct(rt reflect.Type, field reflect.StructField, ns string, index []int) error {
	ns = namespace(field, ns)
	for i := 0; i < rt.NumField(); i++ {
		if err := sv.deepTraverse(rt.Field(i).Type, rt.Field(i), ns, append(index, i)); err != nil {
			return err
		}
	}
	return nil
}

func (sv *structType) traverseField(rt reflect.Type, field reflect.StructField, ns string, index []int) error {
	ns = namespace(field, ns)

	// special case, json tag
	if _, ok := field.Tag.Lookup("json"); ok {
		sv.hasJSONTag = true
	}

	// err tag
	if v, ok := field.Tag.Lookup(sv.g.errTagName); ok {
		sv.errMap[ns] = v
	}

	fInfo := &fieldInfo{
		namespace:   ns,
		index:       index,
		structField: field,
		excer:       nil,
		defaultOpt: DefaultOption{
			DefaultSplitFlag: sv.g.defaultSplitFlag,
		},
	}

	sv.fields[ns] = fInfo

	bindTag, ok := field.Tag.Lookup(sv.g.bindTagName)
	if !ok {
		return nil
	}

	// default tag
	bindTagValue := defaultOpt(bindTag, &fInfo.defaultOpt)

	// excer
	excer, err := sv.g.tagExcers.GetExecer(StringToSlice(bindTagValue))
	if err != nil {
		return nil
	}
	fInfo.excer = excer
	return nil
}

func (sv *structType) parseJSON(data interface{}, obj interface{}) error {
	req, ok := data.(*http.Request)
	if !ok || req == nil || req.Body == nil {
		return e("invalid request")
	}
	decoder := json.NewDecoder(req.Body)
	if sv.g.useNumberForJSON {
		decoder.UseNumber()
	}
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

func head(str, sep string) (head string, tail string) {
	idx := strings.Index(str, sep)
	if idx < 0 {
		return str, ""
	}
	return str[:idx], str[idx+len(sep):]
}

func defaultOpt(tag string, opt *DefaultOption) string {
	excerTag, tail := head(tag, ",")
	var left string
	for len(tail) > 0 {
		left, tail = head(tail, ",")
		if k, v := head(left, "="); k == "default" {
			opt.IsDefaultExists = true
			opt.DefaultValue = v
		}
	}
	return excerTag
}

func namespace(field reflect.StructField, ns string) string {
	if field.Name != "" {
		ns = ns + "." + field.Name
	}
	return ns
}

func deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func e(format string, args ...interface{}) error {
	return fmt.Errorf("gbind: "+format, args...)
}
