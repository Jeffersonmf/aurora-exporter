package util

import (
	"context"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

func UUIDGenerate() string {
	newUUID, err := exec.Command("uuidgen").Output()
	if err != nil {
		Sugar.Errorf("UUIDGenerate Error", zap.String("msg", err.Error()))
	}

	return string(newUUID)
}

// any potentially blocking task should take a context
// style: context should be the first passed in parameter
func PipesTask(ctx context.Context, poll time.Duration) {
	Sugar.Infof("Running...")
	select {
	case <-ctx.Done():
		Sugar.Warnf("PipeTask Execution", zap.String("msg", "Done;"))
		break
	default:
		for {
			time.Sleep(poll * time.Millisecond)
		}
	}
}
