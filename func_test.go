package gomocker

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGen(t *testing.T) {
	err := GenerateMocker(strings.NewReader(`
package dummypackage

import (
	"fmt"
	"context"
)

const (
    x int = iota
    y
    z
)

type Arifin func(ctx context.Context, a, b int) (int, error)
type Finda func(context.Context, [5]int, []int) (int, error)
type Ilmania func(context.Context, int, ...int) (int, error)

type Jauhar interface {
	Integer(ctx context.Context, a int) (int, error)
	Interface(a interface{}, b interface{ MethodA() error }) fmt.Stringer
	Functions(a func()) func(a int, b int) error
}

func main() {
}
`), "Arifin", os.Stdout)
	fmt.Println(err)
}
