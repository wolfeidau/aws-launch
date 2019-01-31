package launcher

import (
	"time"

	"github.com/pkg/errors"
	"github.com/wolfeidau/fargate-run-job/pkg/valid"
)

var (
	// ErrMissingParams missing the params required by the ecs launch
	ErrMissingParams = errors.New("service params are missing from Definition, configure either ECS or Codebuild")
	// ErrInvalidParams either missing or configured more than service one parameters entry
	ErrInvalidParams = errors.New("Requires only one service parameters entry, ecs or codebuild")
)

// Launcher build the definition, then launch a container based task
type Launcher interface {
	CreateDefinition(*DefinitionParams) (*CreateDefinitionResult, error)
	RunTask(*RunTaskParams) (*RunTaskResult, error)
	RunTaskAsync(*RunTaskAsyncParams) (*RunTaskAsyncResult, error)
	GetTaskStatus(*GetTaskStatusParams) (*GetTaskStatusResult, error)
	WaitForTask(*WaitForTaskParams) (*WaitForTaskResult, error)
}

// BaseTaskResult common base task result
type BaseTaskResult struct {
	ECS        *RunTaskECSResult
	CodeBuild  *RunTaskCodebuildResult
	ID         string
	Successful bool
	StartTime  *time.Time
	EndTime    *time.Time
}

// RunTaskECSResult ecs related result information
type RunTaskECSResult struct {
	TaskArn string
	TaskID  string
}

// RunTaskCodebuildResult codebuild related result information
type RunTaskCodebuildResult struct {
	BuildArn    string
	BuildStatus string
}

// Valid validate input structure of run task params
func (rt *BaseTaskResult) Valid() error {
	// do we have any service params at all
	if valid.CountOfNotNil(rt.ECS, rt.CodeBuild) == 0 {
		return ErrMissingParams
	}
	// check there is only one service configuration supplied
	if valid.OneOf(rt.ECS, rt.CodeBuild) {
		return ErrInvalidParams
	}

	return nil
}

// GetTaskStatusParams get status task parameters
type GetTaskStatusParams struct {
	ID        string
	ECS       *ECSTaskParams
	Codebuild *CodebuildTaskParams
}

// GetTaskStatusResult get status task result
type GetTaskStatusResult struct {
	*BaseTaskResult
}

// WaitForTaskParams wait for task parameters
type WaitForTaskParams struct {
	ID        string
	ECS       *ECSTaskParams       `json:"ecs,omitempty"`
	Codebuild *CodebuildTaskParams `json:"codebuild,omitempty"`
}

// WaitForTaskResult wait for task parameters
type WaitForTaskResult struct {
	ID string
}

// RunTaskResult summarsied result of the launched task
type RunTaskResult struct {
	*BaseTaskResult
}

// RunTaskAsyncResult summarsied result of the launched task
type RunTaskAsyncResult struct {
	*BaseTaskResult
}

// RunTaskParams used to launch container based tasks
type RunTaskParams struct {
	ECS         *ECSTaskParams       `json:"ecs,omitempty"`
	Codebuild   *CodebuildTaskParams `json:"codebuild,omitempty"`
	Environment map[string]string    `json:"environment,omitempty"`
	Tags        map[string]string    `json:"tags,omitempty"`
}

// Valid validate input structure of run task params
func (rt *RunTaskParams) Valid() error {
	// do we have any service params at all
	if valid.CountOfNotNil(rt.ECS, rt.Codebuild) == 0 {
		return ErrMissingParams
	}
	// check there is only one service configuration supplied
	if valid.OneOf(rt.ECS, rt.Codebuild) {
		return ErrInvalidParams
	}

	return nil
}

// RunTaskAsyncParams used to launch container based tasks
type RunTaskAsyncParams struct {
	ECS         *ECSTaskParams       `json:"ecs,omitempty"`
	Codebuild   *CodebuildTaskParams `json:"codebuild,omitempty"`
	Environment map[string]string    `json:"environment,omitempty"`
	Tags        map[string]string    `json:"tags,omitempty"`
}

// Valid validate input structure of run task params
func (rt *RunTaskAsyncParams) Valid() error {
	// do we have any service params at all
	if valid.CountOfNotNil(rt.ECS, rt.Codebuild) == 0 {
		return ErrMissingParams
	}
	// check there is only one service configuration supplied
	if valid.OneOf(rt.ECS, rt.Codebuild) {
		return ErrInvalidParams
	}

	return nil
}

// ECSTaskParams ECS related run task parameters
type ECSTaskParams struct {
	ClusterName    string   `json:"cluster_name,omitempty" jsonschema:"required"`
	ServiceName    string   `json:"service_name,omitempty" jsonschema:"required"`
	ContainerName  string   `json:"container_name,omitempty" jsonschema:"required"`
	TaskDefinition string   `json:"task_definition,omitempty" jsonschema:"required"`
	CPU            int64    `json:"cpu,omitempty" jsonschema:"required"`
	Memory         int64    `json:"memory,omitempty" jsonschema:"required"`
	Subnets        []string `json:"subnets,omitempty" jsonschema:"required"`
}

// CodebuildTaskParams Codebuild related run task parameters
type CodebuildTaskParams struct {
	ProjectName    string  `json:"project_name,omitempty" jsonschema:"required"`
	ComputeType    *string `json:"compute_type,omitempty"`
	PrivilegedMode *bool   `json:"privileged_mode,omitempty"`
	Image          *string `json:"image,omitempty"`
	ServiceRole    *string `json:"service_role,omitempty"`
}

// ECSDefinitionParams ECS related definition parameters
type ECSDefinitionParams struct {
	ExecutionRoleARN string `json:"execution_role_arn,omitempty" jsonschema:"required"`
	DefinitionName   string `json:"definition_name,omitempty" jsonschema:"required"`
	ContainerName    string `json:"container_name,omitempty" jsonschema:"required"`
}

// CodebuildDefinitionParams Codebuild related definition parameters
type CodebuildDefinitionParams struct {
	ProjectName    string `json:"project_name,omitempty" jsonschema:"required"`
	ComputeType    string `json:"compute_type,omitempty" jsonschema:"required"`
	PrivilegedMode *bool  `json:"privileged_mode,omitempty"`
	Buildspec      string `json:"buildspec,omitempty" jsonschema:"required"`
	ServiceRole    string `json:"service_role,omitempty" jsonschema:"required"`
}

// DefinitionParams parameters used to build a container execution environment
type DefinitionParams struct {
	ECS         *ECSDefinitionParams       `json:"ecs,omitempty"`
	Codebuild   *CodebuildDefinitionParams `json:"codebuild,omitempty"`
	Region      string                     `json:"region,omitempty" jsonschema:"required"`
	TaskRoleARN *string                    `json:"task_role_arn,omitempty"` // optional
	Image       string                     `json:"image,omitempty" jsonschema:"required"`
	Environment map[string]string          `json:"environment,omitempty"`
	Tags        map[string]string          `json:"tags,omitempty"`
}

// CreateDefinitionResult the results from create definition
type CreateDefinitionResult struct {
	ID string
}

// Valid validate input structure of definition params
func (dp *DefinitionParams) Valid() error {
	// do we have any service params at all
	if valid.CountOfNotNil(dp.ECS, dp.Codebuild) == 0 {
		return ErrMissingParams
	}
	// check there is only one service configuration supplied
	if valid.OneOf(dp.ECS, dp.Codebuild) {
		return ErrInvalidParams
	}

	return nil
}
