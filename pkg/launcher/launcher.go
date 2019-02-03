package launcher

import (
	"time"

	"github.com/pkg/errors"
	"github.com/wolfeidau/fargate-run-job/pkg/valid"
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

// Launcher build the definition, then launch a container based task
type Launcher interface {
	DefineTask(*DefineTaskParams) (*DefineTaskResult, error)
	LaunchTask(*LaunchTaskParams) (*LaunchTaskResult, error)
	DefineAndLaunch(*DefineAndLaunchParams) (*DefineAndLaunchResult, error)
	GetTaskStatus(*GetTaskStatusParams) (*GetTaskStatusResult, error)
	WaitForTask(*WaitForTaskParams) (*WaitForTaskResult, error)
	CleanupTask(*CleanupTaskParams) (*CleanupTaskResult, error)
}

// DefineAndLaunchParams define and launch parameters
type DefineAndLaunchParams struct {
	ECS         *ECSDefineAndLaunchParams       `json:"ecs,omitempty"`
	Codebuild   *CodebuildDefineAndLaunchParams `json:"codebuild,omitempty"`
	Region      string                          `json:"region,omitempty" jsonschema:"required"`
	TaskRoleARN *string                         `json:"task_role_arn,omitempty"` // optional
	Image       string                          `json:"image,omitempty" jsonschema:"required"`
	Environment map[string]string               `json:"environment,omitempty"`
	Tags        map[string]string               `json:"tags,omitempty"`
}

// Valid validate input structure of define and launch task params
func (dlp *DefineAndLaunchParams) Valid() error {
	// do we have any service params at all
	if valid.CountOfNotNil(dlp.ECS, dlp.Codebuild) == 0 {
		return ErrMissingParams
	}
	// check there is only one service configuration supplied
	if valid.OneOf(dlp.ECS, dlp.Codebuild) {
		return ErrInvalidParams
	}

	return nil
}

// BuildDefineTask build a define task params
func (dlp *DefineAndLaunchParams) BuildDefineTask() *DefineTaskParams {

	dt := &DefineTaskParams{
		Region:      dlp.Region,
		TaskRoleARN: dlp.TaskRoleARN,
		Image:       dlp.Image,
		Environment: dlp.Environment,
		Tags:        dlp.Tags,
	}

	if dlp.Codebuild != nil {
		dt.Codebuild = &CodebuildDefineTaskParams{
			ProjectName:    dlp.Codebuild.ProjectName,
			ComputeType:    dlp.Codebuild.ComputeType,
			PrivilegedMode: dlp.Codebuild.PrivilegedMode,
			Buildspec:      dlp.Codebuild.Buildspec,
			ServiceRole:    dlp.Codebuild.ServiceRole,
		}
	}
	if dlp.ECS != nil {
		name := dlp.ECS.DefinitionName

		dt.ECS = &ECSDefineTaskParams{
			ExecutionRoleARN: dlp.ECS.ExecutionRoleARN,
			DefinitionName:   dlp.ECS.DefinitionName,
			ContainerName:    name,
		}
	}
	return dt
}

// BuildLaunchTask build a launch task params
func (dlp *DefineAndLaunchParams) BuildLaunchTask(definitionID string) *LaunchTaskParams {
	lt := &LaunchTaskParams{
		Environment: dlp.Environment,
		Tags:        dlp.Tags,
	}

	if dlp.Codebuild != nil {
		lt.Codebuild = &CodebuildTaskParams{
			ProjectName: dlp.Codebuild.ProjectName,
		}
	}

	if dlp.ECS != nil {
		name := dlp.ECS.DefinitionName

		lt.ECS = &ECSTaskParams{
			ClusterName:    dlp.ECS.ClusterName,
			ServiceName:    name,
			ContainerName:  name,
			TaskDefinition: definitionID,
			CPU:            dlp.ECS.CPU,
			Memory:         dlp.ECS.Memory,
			Subnets:        dlp.ECS.Subnets,
		}
	}

	return lt
}

// DefineAndLaunchResult define and launch result
type DefineAndLaunchResult struct {
	*BaseTaskResult
	CloudwatchLogGroupName string `json:"cloudwatch_log_group_name,omitempty"`
	CloudwatchStreamPrefix string `json:"cloudwatch_stream_prefix,omitempty"`
	DefinitionID           string `json:"definition_id,omitempty"`
}

// ECSDefineAndLaunchParams define and launch parameters for ecs
type ECSDefineAndLaunchParams struct {
	DefinitionName   string   `json:"definition_name,omitempty" jsonschema:"required"`
	ClusterName      string   `json:"cluster_name,omitempty" jsonschema:"required"`
	CPU              int64    `json:"cpu,omitempty" jsonschema:"required"`
	Memory           int64    `json:"memory,omitempty" jsonschema:"required"`
	Subnets          []string `json:"subnets,omitempty" jsonschema:"required"`
	ExecutionRoleARN string   `json:"execution_role_arn,omitempty" jsonschema:"required"`
}

// CodebuildDefineAndLaunchParams define and launch parameters for Codebuild
type CodebuildDefineAndLaunchParams struct {
	ProjectName    string `json:"project_name,omitempty" jsonschema:"required"`
	ComputeType    string `json:"compute_type,omitempty" jsonschema:"required"`
	PrivilegedMode *bool  `json:"privileged_mode,omitempty"`
	Buildspec      string `json:"buildspec,omitempty" jsonschema:"required"`
	ServiceRole    string `json:"service_role,omitempty" jsonschema:"required"`
}

// BaseTaskResult common base task result
type BaseTaskResult struct {
	ECS        *LaunchTaskECSResult
	CodeBuild  *LaunchTaskCodebuildResult
	ID         string
	TaskStatus string
	StartTime  *time.Time
	EndTime    *time.Time
}

// LaunchTaskECSResult ecs related result information
type LaunchTaskECSResult struct {
	TaskArn string
	TaskID  string
}

// LaunchTaskCodebuildResult codebuild related result information
type LaunchTaskCodebuildResult struct {
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

// LaunchTaskResult summarsied result of the launched task
type LaunchTaskResult struct {
	*BaseTaskResult
}

// LaunchTaskParams used to launch container based tasks
type LaunchTaskParams struct {
	ECS         *ECSTaskParams       `json:"ecs,omitempty"`
	Codebuild   *CodebuildTaskParams `json:"codebuild,omitempty"`
	Environment map[string]string    `json:"environment,omitempty"`
	Tags        map[string]string    `json:"tags,omitempty"`
}

// Valid validate input structure of run task params
func (rt *LaunchTaskParams) Valid() error {
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

// ECSDefineTaskParams ECS related definition parameters
type ECSDefineTaskParams struct {
	ExecutionRoleARN string `json:"execution_role_arn,omitempty" jsonschema:"required"`
	DefinitionName   string `json:"definition_name,omitempty" jsonschema:"required"`
	ContainerName    string `json:"container_name,omitempty" jsonschema:"required"`
}

// CodebuildDefineTaskParams Codebuild related definition parameters
type CodebuildDefineTaskParams struct {
	ProjectName    string `json:"project_name,omitempty" jsonschema:"required"`
	ComputeType    string `json:"compute_type,omitempty" jsonschema:"required"`
	PrivilegedMode *bool  `json:"privileged_mode,omitempty"`
	Buildspec      string `json:"buildspec,omitempty" jsonschema:"required"`
	ServiceRole    string `json:"service_role,omitempty" jsonschema:"required"`
}

// DefineTaskParams parameters used to build a container execution environment
type DefineTaskParams struct {
	ECS         *ECSDefineTaskParams       `json:"ecs,omitempty"`
	Codebuild   *CodebuildDefineTaskParams `json:"codebuild,omitempty"`
	Region      string                     `json:"region,omitempty" jsonschema:"required"`
	TaskRoleARN *string                    `json:"task_role_arn,omitempty"` // optional
	Image       string                     `json:"image,omitempty" jsonschema:"required"`
	Environment map[string]string          `json:"environment,omitempty"`
	Tags        map[string]string          `json:"tags,omitempty"`
}

// DefineTaskResult the results from create definition
type DefineTaskResult struct {
	ID                     string `json:"id,omitempty"`
	CloudwatchLogGroupName string `json:"cloudwatch_log_group_name,omitempty"`
	CloudwatchStreamPrefix string `json:"cloudwatch_stream_prefix,omitempty"`
}

// Valid validate input structure of definition params
func (dp *DefineTaskParams) Valid() error {
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

// CleanupTaskParams cleanup definition params
type CleanupTaskParams struct {
	ECS       *ECSCleanupTaskParams       `json:"ecs,omitempty"`
	Codebuild *CodebuildCleanupTaskParams `json:"codebuild,omitempty"`
}

// Valid validate input structure of definition params
func (dp *CleanupTaskParams) Valid() error {
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

// ECSCleanupTaskParams cleanup definition params for ecs
type ECSCleanupTaskParams struct {
	TaskDefinition string `json:"task_definition,omitempty" jsonschema:"required"`
}

// CodebuildCleanupTaskParams cleanup definition params for codebuild
type CodebuildCleanupTaskParams struct {
	ProjectName string `json:"project_name,omitempty" jsonschema:"required"`
}

// CleanupTaskResult cleanup definition result
type CleanupTaskResult struct {
}
