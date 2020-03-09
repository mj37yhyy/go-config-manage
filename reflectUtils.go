package go_config_manage

import "reflect"

func GetInterface(obj interface{}) (interface{}, reflect.Value) {
	targetActual := reflect.ValueOf(obj).Elem()
	configType := targetActual.Type()
	baseReflect := reflect.New(configType)
	// Actual type.
	base := baseReflect.Elem().Interface()
	return base, baseReflect
}
