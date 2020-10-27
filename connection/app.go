package connection

import (
	"context"
)

type hookFn func(context.Context)

//------------------------------------------------------------------------------

var (
	primarily []hookFn
	secondary []hookFn
)

func OnExit(h hookFn) {
	primarily = append(primarily, h)
}

func OnExitSecondary(h hookFn) {
	secondary = append(secondary, h)
}
