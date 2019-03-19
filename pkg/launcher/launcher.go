package launcher

import (
	"github.com/pkg/errors"
)

const (
	// TaskRunning task failed
	TaskRunning = "RUNNING"
	// TaskFailed task failed
	TaskFailed = "FAILED"
	// TaskStopped task stopped
	TaskStopped = "STOPPED"
	// TaskSucceeded task succeeded
	TaskSucceeded = "SUCCEEDED"
)

var (
	// ErrMissingParams missing the params required by the ecs launch
	ErrMissingParams = errors.New("service params are missing from Definition, configure either ECS or Codebuild")
	// ErrInvalidParams either missing or configured more than service one parameters entry
	ErrInvalidParams = errors.New("Requires only one service parameters entry, ecs or codebuild")
)
