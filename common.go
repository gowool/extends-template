package et

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"regexp"
	"unsafe"
)

const (
	extendsPattern  = `%s\s*extends\s*"(.*?)"\s*%s`
	templatePattern = `%s.*?template\s*"(.*?)".*?%s`
)

func ReExtends(left, right string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(extendsPattern, left, right))
}

func ReTemplate(left, right string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(templatePattern, left, right))
}

func merge(left error, right error) error {
	switch {
	case left == nil:
		return right
	case right == nil:
		return left
	}

	return fmt.Errorf("%w; %w", left, right)
}

func errorf(right error, format string, a ...any) error {
	return merge(fmt.Errorf(format, a...), right)
}

func typeName(i any) string {
	t := reflect.TypeOf(i)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}

func toBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func toString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func hash(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
