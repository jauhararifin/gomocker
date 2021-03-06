package examples

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMathMocked_Add(t *testing.T) {
	t.Run("mock return values once", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.MockOutputsOnce(10, nil)
		sum, err := math.Add(context.Background(), 17, 19)
		assert.Nil(t, err, "wrong error")
		assert.Equal(t, 10, sum, "wrong sum")
		assert.Equal(t, 1, len(mocker.Add.Invocations()))
		assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 10, nil)
	})

	t.Run("mock return values 10 times", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.MockOutputs(10, 10, nil)
		for i := 0; i < 10; i++ {
			sum, err := math.Add(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 10, sum, "wrong sum")
		}
		assert.Equal(t, 10, len(mocker.Add.Invocations()))
		for i := 0; i < 10; i++ {
			assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 10, nil)
		}
	})

	t.Run("mock return values forever", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.MockOutputsForever(10, nil)
		for i := 0; i < 100; i++ {
			sum, err := math.Add(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 10, sum, "wrong sum")
		}
		assert.Equal(t, 100, len(mocker.Add.Invocations()))
		for i := 0; i < 100; i++ {
			assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 10, nil)
		}
	})

	t.Run("mock default return values once", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.MockDefaultsOnce()
		sum, err := math.Add(context.Background(), 17, 19)
		assert.Nil(t, err, "wrong error")
		assert.Equal(t, 0, sum, "wrong sum")
		assert.Equal(t, 1, len(mocker.Add.Invocations()))
		assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 0, nil)
	})

	t.Run("mock default return values 10 times", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.MockDefaults(10)
		for i := 0; i < 10; i++ {
			sum, err := math.Add(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 0, sum, "wrong sum")
		}
		assert.Equal(t, 10, len(mocker.Add.Invocations()))
		for i := 0; i < 10; i++ {
			assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 0, nil)
		}
	})

	t.Run("mock default return values forever", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.MockDefaultsForever()
		for i := 0; i < 100; i++ {
			sum, err := math.Add(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 0, sum, "wrong sum")
		}
		assert.Equal(t, 100, len(mocker.Add.Invocations()))
		for i := 0; i < 100; i++ {
			assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 0, nil)
		}
	})

	t.Run("mock func once", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.MockOnce(func(ctx context.Context, a int, b int) (sum int, err error) {
			return a + b, nil
		})
		sum, err := math.Add(context.Background(), 17, 19)
		assert.Nil(t, err, "wrong error")
		assert.Equal(t, 36, sum, "wrong sum")
		assert.Equal(t, 1, len(mocker.Add.Invocations()))
		assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 36, nil)
	})

	t.Run("mock func 10 times", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.Mock(10, func(ctx context.Context, a int, b int) (sum int, err error) {
			return a + b, nil
		})
		for i := 0; i < 10; i++ {
			sum, err := math.Add(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 36, sum, "wrong sum")
		}
		assert.Equal(t, 10, len(mocker.Add.Invocations()))
		for i := 0; i < 10; i++ {
			assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 36, nil)
		}
	})

	t.Run("mock func forever", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Add.MockForever(func(ctx context.Context, a int, b int) (sum int, err error) {
			return a + b, nil
		})
		for i := 0; i < 100; i++ {
			sum, err := math.Add(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 36, sum, "wrong sum")
		}
		assert.Equal(t, 100, len(mocker.Add.Invocations()))
		for i := 0; i < 100; i++ {
			assertMathAddOneInvocation(t, mocker, context.Background(), 17, 19, 36, nil)
		}
	})
}

func assertMathAddOneInvocation(t *testing.T, mocker *MathMocker, ctx context.Context, a, b, sum int, err error) {
	invoc := mocker.Add.TakeOneInvocation()
	assert.Equal(t, ctx, invoc.Inputs.Ctx, "wrong context")
	assert.Equal(t, a, invoc.Inputs.A, "wrong first integer")
	assert.Equal(t, b, invoc.Inputs.B, "wrong second integer")
	assert.Equal(t, sum, invoc.Outputs.Sum, "wrong first returns")
	assert.Equal(t, err, invoc.Outputs.Err, "wrong second returns")
}

func TestMathMocked_Subtract(t *testing.T) {
	t.Run("mock return values once", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.MockOutputsOnce(10, nil)
		sum, err := math.Subtract(context.Background(), 17, 19)
		assert.Nil(t, err, "wrong error")
		assert.Equal(t, 10, sum, "wrong result")
		assert.Equal(t, 1, len(mocker.Subtract.Invocations()))
		assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 10, nil)
	})

	t.Run("mock return values 10 times", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.MockOutputs(10, 10, nil)
		for i := 0; i < 10; i++ {
			sum, err := math.Subtract(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 10, sum, "wrong sum")
		}
		assert.Equal(t, 10, len(mocker.Subtract.Invocations()))
		for i := 0; i < 10; i++ {
			assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 10, nil)
		}
	})

	t.Run("mock return values forever", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.MockOutputsForever(10, nil)
		for i := 0; i < 100; i++ {
			sum, err := math.Subtract(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 10, sum, "wrong sum")
		}
		assert.Equal(t, 100, len(mocker.Subtract.Invocations()))
		for i := 0; i < 100; i++ {
			assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 10, nil)
		}
	})

	t.Run("mock default return values once", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.MockDefaultsOnce()
		sum, err := math.Subtract(context.Background(), 17, 19)
		assert.Nil(t, err, "wrong error")
		assert.Equal(t, 0, sum, "wrong sum")
		assert.Equal(t, 1, len(mocker.Subtract.Invocations()))
		assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 0, nil)
	})

	t.Run("mock default return values 10 times", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.MockDefaults(10)
		for i := 0; i < 10; i++ {
			sum, err := math.Subtract(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 0, sum, "wrong sum")
		}
		assert.Equal(t, 10, len(mocker.Subtract.Invocations()))
		for i := 0; i < 10; i++ {
			assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 0, nil)
		}
	})

	t.Run("mock default return values forever", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.MockDefaultsForever()
		for i := 0; i < 100; i++ {
			sum, err := math.Subtract(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 0, sum, "wrong sum")
		}
		assert.Equal(t, 100, len(mocker.Subtract.Invocations()))
		for i := 0; i < 100; i++ {
			assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 0, nil)
		}
	})

	t.Run("mock func once", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.MockOnce(func(ctx context.Context, a int, b int) (sum int, err error) {
			return a + b, nil
		})
		sum, err := math.Subtract(context.Background(), 17, 19)
		assert.Nil(t, err, "wrong error")
		assert.Equal(t, 36, sum, "wrong sum")
		assert.Equal(t, 1, len(mocker.Subtract.Invocations()))
		assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 36, nil)
	})

	t.Run("mock func 10 times", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.Mock(10, func(ctx context.Context, a int, b int) (sum int, err error) {
			return a + b, nil
		})
		for i := 0; i < 10; i++ {
			sum, err := math.Subtract(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 36, sum, "wrong sum")
		}
		assert.Equal(t, 10, len(mocker.Subtract.Invocations()))
		for i := 0; i < 10; i++ {
			assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 36, nil)
		}
	})

	t.Run("mock func forever", func(t *testing.T) {
		math, mocker := NewMockedMath()
		mocker.Subtract.MockForever(func(ctx context.Context, a int, b int) (sum int, err error) {
			return a + b, nil
		})
		for i := 0; i < 100; i++ {
			sum, err := math.Subtract(context.Background(), 17, 19)
			assert.Nil(t, err, "wrong error")
			assert.Equal(t, 36, sum, "wrong sum")
		}
		assert.Equal(t, 100, len(mocker.Subtract.Invocations()))
		for i := 0; i < 100; i++ {
			assertMathSubtractOneInvocation(t, mocker, context.Background(), 17, 19, 36, nil)
		}
	})
}

func assertMathSubtractOneInvocation(t *testing.T, mocker *MathMocker, ctx context.Context, a, b, result int, err error) {
	invoc := mocker.Subtract.TakeOneInvocation()
	assert.Equal(t, ctx, invoc.Inputs.Ctx, "wrong context")
	assert.Equal(t, a, invoc.Inputs.A, "wrong first integer")
	assert.Equal(t, b, invoc.Inputs.B, "wrong second integer")
	assert.Equal(t, result, invoc.Outputs.Result, "wrong first returns")
	assert.Equal(t, err, invoc.Outputs.Err, "wrong second returns")
}
