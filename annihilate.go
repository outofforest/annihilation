package annihilation

import (
	"context"

	"github.com/pkg/errors"

	"github.com/outofforest/parallel"
)

// AppFunc is a function running application.
type AppFunc func(ctx context.Context, args []string) error

// Apps is a set of defined apps.
type Apps map[string]AppFunc

// Run runs applications given CLI arguments.
func Run(apps Apps, args ...string) ([]string, parallel.Task, error) {
	if len(args) < 2 {
		return nil, nil, errors.New("no arguments provided")
	}

	parsedApps := make([]appInfo, 0, len(apps))

	var info appInfo
	for _, arg := range args[1:] {
		if appFunc, exists := apps[arg]; exists {
			parsedApps = append(parsedApps, info)
			info = appInfo{
				Name: arg,
				Func: appFunc,
			}
			continue
		}
		info.Args = append(info.Args, arg)
	}
	parsedApps = append(parsedApps, info)

	if len(parsedApps) < 2 {
		return nil, nil, errors.New("no applications to execute")
	}

	localArgs := parsedApps[0].Args
	parsedApps = parsedApps[1:]

	return localArgs, func(ctx context.Context) error {
		if len(parsedApps) == 1 {
			return parsedApps[0].Func(ctx, parsedApps[0].Args)
		}

		return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
			for _, app := range parsedApps {
				spawn(app.Name, parallel.Fail, func(ctx context.Context) error {
					return app.Func(ctx, app.Args)
				})
			}
			return nil
		})
	}, nil
}

type appInfo struct {
	Name string
	Args []string
	Func AppFunc
}
