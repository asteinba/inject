package inject

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

func NewInjector() *Injector {
	return &Injector{
		provider: make(map[interface{}]reflect.Value, 0),
	}
}

type Injector struct {
	provider map[interface{}]reflect.Value
	mutex    sync.RWMutex
}

// Provide registers the given dependencies.
func (inj *Injector) Provide(deps ...interface{}) {
	for i := 0; i < len(deps); i++ {
		valueOfProvider := reflect.ValueOf(deps[i])
		inj.mutex.Lock()
		inj.provider[valueOfProvider.Type()] = valueOfProvider
		inj.mutex.Unlock()
	}
}

// ProvideNamed registers the given dependency with the given name.
func (inj *Injector) ProvideNamed(dep interface{}, name string) {
	inj.mutex.Lock()
	inj.provider[name] = reflect.ValueOf(dep)
	inj.mutex.Unlock()
}

// Inject dependencies into dst. The argument dst needs to be pointer to a struct.
// The corresponding struct needs to define a struct tag named "inject".
// If the struct tag is empty (inject:"") than the dependency is resolved using the type of the field.
// You can pass a optional name to the struct tag (`inject:"name"`) and then the Injector only resolves the dependency
// using the optional provider name of the Injector.Provide func. This is helpful if you want to pass multiple
// dependencies with the same type.
func (inj Injector) Inject(dst interface{}) error {
	valueOfDst := reflect.ValueOf(dst)
	if valueOfDst.Kind() != reflect.Ptr {
		return errors.New(`Argument module has to be a pointer to a struct.`)
	}

	valueOfDst = reflect.Indirect(valueOfDst)
	if valueOfDst.Kind() != reflect.Struct {
		return errors.New(`Argument module has to be a pointer to a struct.`)
	}
	typeOfDst := valueOfDst.Type()

	for i := 0; i < typeOfDst.NumField(); i++ {
		valueField := valueOfDst.Field(i)
		typeField := typeOfDst.Field(i)
		name, ok := typeField.Tag.Lookup("inject")
		if !ok {
			continue
		}

		var dep reflect.Value
		inj.mutex.RLock()
		if name != "" {
			dep, ok = inj.provider[name]
		} else {
			dep, ok = inj.provider[typeField.Type]
		}
		inj.mutex.RUnlock()

		if !ok {
			return fmt.Errorf(`Missing dependency of type %v.`, typeField.Type.String())
		}

		valueField.Set(dep)
	}
	return nil
}

// This func is only a wrapper for Injector.Inject which panics if Injector.Inject returns a error.
func (inj Injector) MustInject(dst interface{}) {
	err := inj.Inject(dst)
	if err != nil {
		panic(err)
	}
}