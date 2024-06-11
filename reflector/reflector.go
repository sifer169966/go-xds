package reflector

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Reflector interface {
	Watch(ctx context.Context) error
}

// Start ... running the Reflector(s) in a separate goroutine and return if got the stop signal or one of them has got an error
func Start(stopCtx context.Context, rf ...Reflector) error {
	g, ctx := errgroup.WithContext(stopCtx)
	for i := 0; i < len(rf); i++ {
		preFun := func(i int) func() error {
			return func() error {
				return rf[i].Watch(ctx)
			}
		}
		fun := preFun(i)
		g.Go(fun)
	}
	return g.Wait()
}
