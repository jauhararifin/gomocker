package examples

type SomeDummyStruct struct {
	A int
	B int
}

type DoSomethingWithSomeDummyStruct func(dummy SomeDummyStruct)
