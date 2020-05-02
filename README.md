# Go Mocker

Gomocker generates Golang code to easily mock a function or interface

## Installing

```bash
go get github.com/jauhararifin/gomocker/gomocker
```

## Example

1. Suppose you have an interface that look like this

```go
type Math interface {
	Add(ctx context.Context, a, b int) (sum int, err error)
	Subtract(ctx context.Context, a, b int) (result int, err error)
}
```

2. Put a go generate script to generate mocker code for that interface
```go
//go:generate gomocker gen --force Math

type Math interface {
	Add(ctx context.Context, a, b int) (sum int, err error)
	Subtract(ctx context.Context, a, b int) (result int, err error)
}
```

3. Run `go generate ./...`

4. A file called `math_mock_gen.go` containing all the code to mock `Math` interface will be generated.

5. Use the mocker

```go
func TestMathMocked_Add(t *testing.T) {
	// create the mocked `Math` and it's mocker.
	math, mocker := NewMockedMath()

	// mock the output value of `Add` method once.
	mocker.Add.MockOutputsOnce(10, nil)

	// call the Add method. 
	sum, err := math.Add(context.Background(), 17, 19)

	// it should return 10 because we mock it previously.
	assert.Nil(t, err, "wrong error")
	assert.Equal(t, 10, sum, "wrong sum")
	
	// use `mocker.Add.TakeOnceInvocation` to get the "invocation struct". The "invocation struct" contains all the
	// input and output of an invocation.
	invoc := mocker.Add.TakeOneInvocation()
	assert.Equal(t, context.Background(), invoc.Inputs.Ctx, "wrong context")
	assert.Equal(t, 17, invoc.Inputs.A, "wrong first integer")
	assert.Equal(t, 19, invoc.Inputs.B, "wrong second integer")
	assert.Equal(t, 10, invoc.Outputs.Sum, "wrong first returns")
	assert.Equal(t, nil, invoc.Outputs.Err, "wrong second returns")
}
```
