package gbind

import (
	"reflect"
	"runtime"
	"unsafe"
)

// ReflectTypeID returns unique type identifier of v.
func ReflectTypeID(v interface{}) uintptr {
	// rt := reflect.TypeOf(v)
	// return uintptr(unsafe.Pointer(rt))
	return 0
}

// MethodName get the name of the current executing function
func MethodName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unkown method"
	}
	return f.Name()
}

// SliceToString slice to string with out data copy
func SliceToString(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}

// StringToSlice string to slice with out data copy
func StringToSlice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}
