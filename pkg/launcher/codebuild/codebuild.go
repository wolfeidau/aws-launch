package codebuild

import (
	"time"

	"github.com/wolfeidau/aws-launch/pkg/cwlogs"
)

const (
	// CodebuildStreamPrefix the prefix used in the cloudwatch log stream name
	CodebuildStreamPrefix = "codebuild"

	// CodebuildLogGroupFormat the name format for cloudwatch log group names
	CodebuildLogGroupFormat = "/aws/codebuild/%s"
)

// LauncherAPI build the definition, then launch a container based task
type LauncherAPI interface {
	DefineTask(*DefineTaskParams) (*DefineTaskResult, error)
	LaunchTask(*LaunchTaskParams) (*LaunchTaskResult, error)
	GetTaskStatus(*GetTaskStatusParams) (*GetTaskStatusResult, error)
	WaitForTask(*WaitForTaskParams) (*WaitForTaskResult, error)
	CleanupTask(*CleanupTaskParams) (*CleanupTaskResult, error)
	GetTaskLogs(*GetTaskLogsParams) (*GetTaskLogsResult, error)
}

// DefineTaskParams parameters used to build a container execution environment for Codebuild
type DefineTaskParams struct {
	ProjectName    string `json:"project_name,omitempty" jsonschema:"required"`
	ComputeType    string `json:"compute_type,omitempty" jsonschema:"required"`
	PrivilegedMode *bool  `json:"privileged_mode,omitempty"`
	Buildspec      string `json:"buildspec,omitempty" jsonschema:"required"`
	ServiceRole    string `json:"service_role,omitempty" jsonschema:"required"`

	Region      string            `json:"region,omitempty" jsonschema:"required"`
	TaskRoleARN *string           `json:"task_role_arn,omitempty"` // optional
	Image       string            `json:"image,omitempty" jsonschema:"required"`
	Environment map[string]string `json:"environment,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// DefineTaskResult the results from create definition for Codebuild
type DefineTaskResult struct {
	ID                     string `json:"id,omitempty"`
	CloudwatchLogGroupName string `json:"cloudwatch_log_group_name,omitempty"`
	CloudwatchStreamPrefix string `json:"cloudwatch_stream_prefix,omitempty"`
}

// LaunchTaskParams used to launch Codebuild container based tasks
type LaunchTaskParams struct {
	ProjectName    string  `json:"project_name,omitempty" jsonschema:"required"`
	ComputeType    *string `json:"compute_type,omitempty"`
	PrivilegedMode *bool   `json:"privileged_mode,omitempty"`
	Image          *string `json:"image,omitempty"`
	ServiceRole    *string `json:"service_role,omitempty"`

	Environment map[string]string `json:"environment,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// LaunchTaskResult summarsied result of the launched task in Codebuild
type LaunchTaskResult struct {
	BuildArn    string
	BuildStatus string

	ID         string
	TaskStatus string
	StartTime  *time.Time
	EndTime    *time.Time
}

// GetTaskStatusParams get status task parameters for Codebuild
type GetTaskStatusParams struct {
	ID string
}

// GetTaskStatusResult get status task result for Codebuild
type GetTaskStatusResult struct {
	BuildArn    string
	BuildStatus string

	ID         string     `json:"id,omitempty"`
	TaskStatus string     `json:"task_status,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
}

// WaitForTaskParams wait for task parameters for Codebuild
type WaitForTaskParams struct {
	ID string `json:"id,omitempty"`
}

// WaitForTaskResult wait for task parameters for Codebuild
type WaitForTaskResult struct {
	ID string
}

// CleanupTaskParams cleanup definition params for Codebuild
type CleanupTaskParams struct {
	ProjectName string `json:"project_name,omitempty" jsonschema:"required"`
}

// CleanupTaskResult cleanup definition result for Codebuild
type CleanupTaskResult struct {
}

// GetTaskLogsParams get logs task params for Codebuild
type GetTaskLogsParams struct {
	ProjectName string  `json:"project_name,omitempty" jsonschema:"required"`
	TaskID      string  `json:"task_id,omitempty" jsonschema:"required"`
	NextToken   *string `json:"next_token,omitempty"`
}

// GetTaskLogsResult get logs task result for Codebuild
type GetTaskLogsResult struct {
	LogLines  []*cwlogs.LogLine `json:"log_lines,omitempty"`
	NextToken *string           `json:"next_token,omitempty"`
}
