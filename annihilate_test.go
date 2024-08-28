package annihilation

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/outofforest/logger"
	"github.com/outofforest/parallel"
)

const (
	box  = "box"
	app1 = "app1"
	app2 = "app2"
)

func newApps() (Apps, <-chan []string, <-chan []string) {
	app1Ch := make(chan []string, 1)
	app2Ch := make(chan []string, 1)
	return Apps{
		app1: func(ctx context.Context, args []string) error {
			<-ctx.Done()
			app1Ch <- args
			return errors.WithStack(ctx.Err())
		},
		app2: func(ctx context.Context, args []string) error {
			<-ctx.Done()
			app2Ch <- args
			return errors.WithStack(ctx.Err())
		},
	}, app1Ch, app2Ch
}

func TestNoArguments(t *testing.T) {
	apps, _, _ := newApps()
	args, runFunc, err := Run(apps, "")
	require.Empty(t, args)
	require.Nil(t, runFunc)
	require.Error(t, err)
}

func TestOneArguments(t *testing.T) {
	apps, _, _ := newApps()
	args, runFunc, err := Run(apps, box)
	require.Empty(t, args)
	require.Nil(t, runFunc)
	require.Error(t, err)
}

func TestNoApps(t *testing.T) {
	apps, _, _ := newApps()
	args, runFunc, err := Run(apps, box, "noApp")
	require.Empty(t, args)
	require.Nil(t, runFunc)
	require.Error(t, err)
}

func TestOneAppNoArguments(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.DefaultConfig))

	apps, app1Ch, app2Ch := newApps()
	args, runFunc, err := Run(apps, box, app1)
	require.NoError(t, err)
	require.Empty(t, args)
	require.NotNil(t, runFunc)

	group := parallel.NewGroup(ctx)
	group.Spawn(box, parallel.Exit, runFunc)
	group.Exit(nil)
	require.NoError(t, group.Wait())

	require.Equal(t, []string{app1}, shouldRun(t, app1Ch))
	shouldNotRun(t, app2Ch)
}

func TestTwoAppsNoArguments(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.DefaultConfig))

	apps, app1Ch, app2Ch := newApps()
	args, runFunc, err := Run(apps, box, app1, app2)
	require.NoError(t, err)
	require.Empty(t, args)
	require.NotNil(t, runFunc)

	group := parallel.NewGroup(ctx)
	group.Spawn(box, parallel.Exit, runFunc)
	group.Exit(nil)
	require.NoError(t, group.Wait())

	require.Equal(t, []string{app1}, shouldRun(t, app1Ch))
	require.Equal(t, []string{app2}, shouldRun(t, app2Ch))
}

func TestTwoAppsWithArguments(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.DefaultConfig))

	apps, app1Ch, app2Ch := newApps()
	args, runFunc, err := Run(apps, box,
		app1,
		"--app1-arg1", "app1-value1",
		"--app1-arg2", "app1-value2",
		app2,
		"--app2-arg1", "app2-value1",
		"--app2-arg2", "app2-value2",
	)
	require.NoError(t, err)
	require.Empty(t, args)
	require.NotNil(t, runFunc)

	group := parallel.NewGroup(ctx)
	group.Spawn(box, parallel.Exit, runFunc)
	group.Exit(nil)
	require.NoError(t, group.Wait())

	require.Equal(t, []string{
		app1,
		"--app1-arg1", "app1-value1",
		"--app1-arg2", "app1-value2",
	}, shouldRun(t, app1Ch))
	require.Equal(t, []string{
		app2,
		"--app2-arg1", "app2-value1",
		"--app2-arg2", "app2-value2",
	}, shouldRun(t, app2Ch))
}

func TestTwoAppsWithArgumentsAndBoxArguments(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.DefaultConfig))

	apps, app1Ch, app2Ch := newApps()
	args, runFunc, err := Run(apps, box,
		"--box-arg1", "box-value1",
		"--box-arg2", "box-value2",
		app1,
		"--app1-arg1", "app1-value1",
		"--app1-arg2", "app1-value2",
		app2,
		"--app2-arg1", "app2-value1",
		"--app2-arg2", "app2-value2",
	)
	require.NoError(t, err)
	require.Equal(t, []string{
		"--box-arg1", "box-value1",
		"--box-arg2", "box-value2",
	}, args)
	require.NotNil(t, runFunc)

	group := parallel.NewGroup(ctx)
	group.Spawn(box, parallel.Exit, runFunc)
	group.Exit(nil)
	require.NoError(t, group.Wait())

	require.Equal(t, []string{
		app1,
		"--app1-arg1", "app1-value1",
		"--app1-arg2", "app1-value2",
	}, shouldRun(t, app1Ch))
	require.Equal(t, []string{
		app2,
		"--app2-arg1", "app2-value1",
		"--app2-arg2", "app2-value2",
	}, shouldRun(t, app2Ch))
}

func shouldRun(t *testing.T, ch <-chan []string) []string {
	select {
	case args := <-ch:
		return args
	default:
		require.Fail(t, "app should run")
		return nil
	}
}

func shouldNotRun(t *testing.T, ch <-chan []string) {
	select {
	case <-ch:
		require.Fail(t, "app should not run")
	default:
	}
}
