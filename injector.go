package inject

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"strings"
)

type ProviderMap map[interface{}]reflect.Value

// Provide registers the given dependencies. This func is NOT safe for concurrent use.
func (pm ProviderMap) Provide(deps ...interface{}) {
	for i := 0; i < len(deps); i++ {
		valueOfProvider := reflect.ValueOf(deps[i])
		pm[valueOfProvider.Type()] = valueOfProvider
	}
}

// ProvideNamed registers the given dependency with the given name. This func is NOT safe for concurrent use.
func (pm ProviderMap) ProvideNamed(dep interface{}, name string) {
	pm[strings.TrimSuffix(name, "*")] = reflect.ValueOf(dep)
}

//Combine this ProviderMap with the given one.
func (pm ProviderMap) combine(src ProviderMap) {
	for k,v := range src{
		pm[k] = v
	}
}

func NewInjector() *Injector {
	return &Injector{
		provider: make(ProviderMap, 0),
	}
}

type Injector struct {
	provider ProviderMap
	mutex    sync.RWMutex
}

// Provide registers the given dependencies. This func is safe for concurrent use.
func (inj *Injector) Provide(deps ...interface{}) {
		inj.mutex.Lock()
		inj.provider.Provide(deps...)
		inj.mutex.Unlock()
}

// ProvideNamed registers the given dependency with the given name. This func is safe for concurrent use.
func (inj *Injector) ProvideNamed(dep interface{}, name string) {
	inj.mutex.Lock()
	inj.provider.ProvideNamed(dep, name)
	inj.mutex.Unlock()
}

// Inject dependencies into dst. The argument dst needs to be pointer to a struct.
// The corresponding struct needs to define a struct tag named "inject".
// If the struct tag is empty (inject:"") than the dependency is resolved using the type of the field.
// You can pass a optional name to the struct tag (`inject:"name"`) and then the Injector only resolves the dependency
// using the optional provider name of the Injector.Provide func. This is helpful if you want to pass multiple
// dependencies with the same type.
// With the argument "extraProviders", you can pass some additional ProviderMaps which are only available while this injection.
func (inj Injector) Inject(dst interface{}, extraProviders ...ProviderMap) error {
	provider := make(ProviderMap)
	inj.mutex.RLock()
	provider.combine(inj.provider)
	inj.mutex.RUnlock()

	for _, extraProv := range extraProviders{
		provider.combine(extraProv)
	}

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

		required := false
		if strings.HasSuffix(name, "*"){
			required = true
			name = strings.TrimSuffix(name, "*")
		}

		var dep reflect.Value
		if name != "" {
			inj.mutex.RLock()
			dep, ok = provider[name]
			inj.mutex.RUnlock()

			if !ok{
				if required{
					return fmt.Errorf(`Missing named "%v" dependency of type %v.`, name, typeField.Type.String())
				}
				continue
			}

			if dep.Type() != valueField.Type() && !dep.Type().Implements(typeField.Type){
				return fmt.Errorf(`Type %v does not fit for field %v of type %v.`, dep.Type(), typeField.Name, typeField.Type)
			}
		} else {
			inj.mutex.RLock()
			dep, ok = provider[typeField.Type]
			inj.mutex.RUnlock()
			if !ok {
				if required {
					return fmt.Errorf(`Missing unnamed dependency of type %v.`, typeField.Type.String())
				}
				continue
			}
		}
		valueField.Set(dep)
	}
	return nil
}

// This func is only a wrapper for Injector.Inject which panics if Injector.Inject returns a error.
func (inj Injector) MustInject(dst interface{}, extraProviders ...ProviderMap) {
	err := inj.Inject(dst, extraProviders...)
	if err != nil {
		panic(err)
	}
}
