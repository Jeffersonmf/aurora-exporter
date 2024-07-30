package core

import (
	"context"

	"brandlovrs.exporter/internal/config"
	workermanager "github.com/Jeffersonmf/go-workers/v3/pkg/worker_manager"
)

var REDSHIFT_CONNECTOR = config.ReadParameter("REDSHIFT_URI")
var AURORA_CONNECTOR = config.ReadParameter("POSTGRESQL_URI")

func init() {

}

func RegisterWorker(ctx context.Context, loadFunc func(context.Context, workermanager.TaskParams) (workermanager.TaskParams, error), writeFunc func(context.Context, workermanager.TaskParams) (workermanager.TaskParams, error)) {

	listenCallbackStep := func(ctx context.Context, taskArg workermanager.TaskParams) error {
		workermanager.Worker{
			SourceContext: ctx,
			Instrumentation: workermanager.Instrumentation{
				TaskArguments:  taskArg,
				FuncDispatcher: writeFunc,
				FuncName:       "Step to Write to Hotdata database",
			},
		}.Run()

		return nil
	}

	workermanager.Worker{
		SourceContext: ctx,
		CronSchedulerConfig: workermanager.CronSchedulerConfig{
			IntervalInSeconds: 0,
			TypeOfExecution:   workermanager.TypeOfExecutionEnum.Async,
		},
		Instrumentation: workermanager.Instrumentation{
			TaskArguments:  workermanager.NewTaskParams(),
			FuncDispatcher: loadFunc,
			NestedCallback: listenCallbackStep,
			FuncName:       "Step to Listen Data from Redshift",
		},
	}.Run()
}
