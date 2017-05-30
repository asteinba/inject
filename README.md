# inject
Small dependency injection package for golang.

See injector_test.go for more examples.

## Package doc
```
type Injector struct {
    // contains filtered or unexported fields
}

func NewInjector() *Injector

func (inj Injector) Inject(dst interface{}) error
    Inject dependencies into dst. The argument dst needs to be pointer to a
    struct. The corresponding struct needs to define a struct tag named
    "inject". If the struct tag is empty (inject:"") than the dependency is
    resolved using the type of the field. You can pass a optional name to
    the struct tag (`inject:"name"`) and then the Injector resolves the
    dependency using the optional provider name of the Injector.Provide
    func. This is helpful if you want to pass multiple dependencies with the
    same type.

func (inj Injector) MustInject(dst interface{})
    This func is only a wrapper for Injector.Inject which panics if
    Injector.Inject returns a error.

func (inj *Injector) Provide(deps ...interface{})
    Provide registers the given dependencies.

func (inj *Injector) ProvideNamed(dep interface{}, name string)
    ProvideNamed registers the given dependency with the given name.

```

## Examples
```
import (
	"fmt"
)

type iNeedSomeDeps struct{
	// Struct fields need to be exported to get resolved.
	// Both below are resolved using the struct type
	A depA `inject:""`
	// Unnamed dependencies with a * in struct tag are required.
	// Inject will throw an error if they are not provided.
	B depB `inject:"*"`
	// Both below are resolved using the name in struct tag
	C dep `inject:"myDepC"`
	// Named dependencies with a trailing * in struct tag are required.
	// Inject will throw an error if they are not provided.
	D dep `inject:"myDepD*"`
}


func main(){
	// Create our injector
	injector := NewInjector()
	
	// Provided without name. Resolved through the type. If a type is provided again then the old is overwritten.
	injector.Provide(depA{prefix: "A"}, depB{prefix: "B"})

	// Provided with name. Duplicate types allowed.
	injector.ProvideNamed(depC{prefix: "C"}, "myDepC")
	injector.ProvideNamed(depD{prefix: "D"}, "myDepD")

	iNeed := iNeedSomeDeps{}
	// Inject the deps into the struct which needs it.
	injector.MustInject(&iNeed)
	
	fmt.Println("A says:", iNeed.A.WhoAmI())
	fmt.Println("B says:", iNeed.B.WhoAmI())
	fmt.Println("D says:", iNeed.C.WhoAmI())
	fmt.Println("C says:", iNeed.D.WhoAmI())
}

type dep interface {
	WhoAmI() string
}

type depA struct {
	prefix string
}

func (a depA) WhoAmI() string {
	return a.prefix+" is my name"
}

type depB struct {
	prefix string
}

func (b depB) WhoAmI() string {
	return b.prefix+" is my name"
}

type depC struct {
	prefix string
}

func (c depC) WhoAmI() string {
	return c.prefix+" is my name"
}

type depD struct {
	prefix string
}

func (d depD) WhoAmI() string {
	return d.prefix+" is my name"
}
```
