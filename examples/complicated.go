package examples

import "context"

type Complicated interface {
	MethodA(ctx context.Context, param1 interface {
		Param1Method1(...int) (int, error)
		Param1Method2(func(int, int, ...int) error) (int, error)
	}, param2 ...struct {
		A, B, C func(ctx2 context.Context, param2 struct {
			A, B, C int
			D, E, F func() interface {
				A() error
			}
		})
	}) (int, int, []int)
	MethodB(chan int, <-chan int, chan<- int, []int, [10]int, map[int]struct{}, map[int]string, map[int]struct {
		A, B, C int
		D, E, F func() interface {
			A() error
		}
	}) (interface {
		Param1Method1(...int) (int, error)
		Param1Method2(func(int, int, ...int) error) (int, error)
	}, struct {
		A, B, C int
		D, E, F func() interface {
			A() error
		}
	})
}
