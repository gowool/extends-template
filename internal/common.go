package internal

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"regexp"
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

func TypeName(i any) string {
	t := reflect.TypeOf(i)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}

func Hash(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
