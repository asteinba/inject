package inject

import (
	"testing"
)

type dep interface {
	WhoAmI() string
}

type depA struct {
	prefix string
}

func (a depA) WhoAmI() string {
	return a.prefix + " is my name"
}

type depB struct {
	prefix string
}

func (b depB) WhoAmI() string {
	return b.prefix + " is my name"
}

type depC struct {
	prefix string
}

func (c depC) WhoAmI() string {
	return c.prefix + " is my name"
}

type depD struct {
	prefix string
}

func (d depD) WhoAmI() string {
	return d.prefix + " is my name"
}

type iNeedSomeDeps struct {
	// Struct fields need to be exported to get resolved.
	// Both below are resolved using the struct type
	A depA `inject:""`
	B depB `inject:""`
	// Both below are resolved using the name in struct tag
	C dep `inject:"myDepC"`
	D dep `inject:"myDepD"`
}


type iNeedSomeDepsMaybeOptional1 struct {
	// Struct fields need to be exported to get resolved.
	// Both below are resolved using the struct type.
	// depB is required because of the *
	A depA `inject:""`
	B depB `inject:"*"`
}

type iNeedSomeDepsMaybeOptional2 struct {
	// Both below are resolved using the name in struct tag
	// myDepD is required because of the trailing *
	C dep `inject:"myDepC"`
	D dep `inject:"myDepD*"`
}

func TestInjector_Inject(t *testing.T) {
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

	if iNeed.A.WhoAmI() != "A is my name" {
		t.Fatal("A says something wrong.")
	}
	if iNeed.B.WhoAmI() != "B is my name" {
		t.Fatal("B says something wrong.")
	}
	if iNeed.C.WhoAmI() != "C is my name" {
		t.Fatal("C says something wrong.")
	}
	if iNeed.D.WhoAmI() != "D is my name" {
		t.Fatal("D says something wrong.")
	}
}

func TestInjector_InjectMissingDeps(t *testing.T) {
	injector := NewInjector()

	injector.Provide(depA{prefix: "A"})
	if injector.Inject(&iNeedSomeDepsMaybeOptional1{}) == nil{
		t.Fatal("Inject should fail because a required unnamed dependency (depB) is missing but required.")
	}

	if injector.Inject(&iNeedSomeDepsMaybeOptional2{}) == nil{
		t.Fatal("Inject should fail because a required named dependency (myDepD) is missing but required.")
	}

	injector.ProvideNamed(depD{prefix: "D"}, "myDepD")
	if err := injector.Inject(&iNeedSomeDepsMaybeOptional2{}); err != nil{
		t.Fatal("Inject should not fail because all required dependencies are provided")
	}
}
