package logger

import (
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

type SecretsRedactor struct {
	replacementsMap map[string]string
}

func NewSecretsRedactor(secrets []string) *SecretsRedactor {
	redactor := &SecretsRedactor{
		replacementsMap: make(map[string]string, len(secrets)),
	}

	for _, secret := range secrets {
		replacement := "**REDACTED**"
		if len(secret) >= 20 {
			replacement = secret[:9] + "..." + secret[len(secret)-5:]
		}
		redactor.replacementsMap[secret] = replacement
	}

	return redactor
}

func (h *SecretsRedactor) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *SecretsRedactor) Fire(entry *logrus.Entry) error {
	entry.Message = h.redactString(entry.Message)

	for key, value := range entry.Data {
		entry.Data[key] = h.redactValue(reflect.ValueOf(value))
	}

	return nil
}

func (h *SecretsRedactor) redactValue(v reflect.Value) any {
	if h.reflectValueIsNil(v) {
		return nil
	}

	if !v.IsValid() {
		return nil
	}

	// TODO: this is far from exhaustive :(
	switch v.Kind() {
	case reflect.String:
		return h.redactString(v.String())

	case reflect.Struct:
		newStruct := reflect.New(v.Type()).Elem()
		h.redactStructFields(v, newStruct)
		return newStruct.Interface()

	case reflect.Slice:
		newSlice := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			newSlice.Index(i).Set(reflect.ValueOf(h.redactValue(v.Index(i))))
		}

		return newSlice.Interface()

	case reflect.Map:
		newMap := reflect.MakeMap(v.Type())
		iter := v.MapRange()
		for iter.Next() {
			newV := h.redactValue(iter.Value())
			newMap.SetMapIndex(iter.Key(), reflect.ValueOf(newV))
		}

		return newMap.Interface()
	}

	return v.Interface()
}

func (h *SecretsRedactor) redactStructFields(src, dest reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		dest.Field(i).Set(reflect.ValueOf(h.redactValue(src.Field(i))))
	}
}

func (h *SecretsRedactor) redactString(s string) string {
	for secret, redacted := range h.replacementsMap {
		s = strings.ReplaceAll(s, secret, redacted)
	}
	return s
}

func (h *SecretsRedactor) reflectValueIsNil(value reflect.Value) bool {
	kind := value.Kind()
	return (kind == reflect.Pointer || kind == reflect.Interface || kind == reflect.Array || kind == reflect.Slice || kind == reflect.Map) && value.IsNil()
}
