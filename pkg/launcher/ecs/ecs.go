package ecs

import (
	"time"

	"github.com/wolfeidau/aws-launch/pkg/cwlogs"
)

const (
	// DefaultCPU default cpu allocation to a fargate task
	DefaultCPU = "256"

	// DefaultMemory default memory allocation to a fargate task
	DefaultMemory = "512"

	// ECSStreamPrefix the prefix used in the ECS cloudwatch log stream name
	ECSStreamPrefix = "ecs"

	// ECSLogGroupFormat the name format for ECS cloudwatch log group names
	ECSLogGroupFormat = "/aws/fargate/%s"
)

// LauncherAPI build the definition, then launch a container based task
type LauncherAPI interface {
	DefineTask(*DefineTaskParams) (*DefineTaskResult, error)
	LaunchTask(*LaunchTaskParams) (*LaunchTaskResult, error)
	GetTaskStatus(*GetTaskStatusParams) (*GetTaskStatusResult, error)
	WaitForTask(*WaitForTaskParams) (*WaitForTaskResult, error)
	StopTask(*StopTaskParams) (*StopTaskResult, error)
	CleanupTask(*CleanupTaskParams) (*CleanupTaskResult, error)
	GetTaskLogs(*GetTaskLogsParams) (*GetTaskLogsResult, error)
}

// DefineTaskParams parameters used to build a container execution environment for Codebuild
type DefineTaskParams struct {
	ExecutionRoleARN string `json:"execution_role_arn,omitempty" jsonschema:"required"`
	DefinitionName   string `json:"definition_name,omitempty" jsonschema:"required"`
	ContainerName    string `json:"container_name,omitempty" jsonschema:"required"`

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
	ClusterName    string   `json:"cluster_name,omitempty" jsonschema:"required"`
	ServiceName    string   `json:"service_name,omitempty" jsonschema:"required"`
	ContainerName  string   `json:"container_name,omitempty" jsonschema:"required"`
	TaskDefinition string   `json:"task_definition,omitempty" jsonschema:"required"`
	CPU            int64    `json:"cpu,omitempty" jsonschema:"required"`
	Memory         int64    `json:"memory,omitempty" jsonschema:"required"`
	Subnets        []string `json:"subnets,omitempty" jsonschema:"required"`

	Environment map[string]string `json:"environment,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// LaunchTaskResult summarsied result of the launched task in Codebuild
type LaunchTaskResult struct {
	TaskArn string `json:"task_arn,omitempty"`
	TaskID  string `json:"task_id,omitempty"`

	ID         string     `json:"id,omitempty"`
	TaskStatus string     `json:"task_status,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
}

// GetTaskStatusParams get status task parameters for Codebuild
type GetTaskStatusParams struct {
	ClusterName string `json:"cluster_name,omitempty" jsonschema:"required"`

	ID string `json:"id,omitempty"`
}

// GetTaskStatusResult get status task result for Codebuild
type GetTaskStatusResult struct {
	TaskArn    string `json:"task_arn,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
	LastStatus string `json:"last_status,omitempty"`
	StopCode   string `json:"stop_reason,omitempty"`

	ID         string     `json:"id,omitempty"`
	TaskStatus string     `json:"task_status,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
}

// WaitForTaskParams wait for task parameters for Codebuild
type WaitForTaskParams struct {
	ClusterName string `json:"cluster_name,omitempty" jsonschema:"required"`

	ID string `json:"id,omitempty"`
}

// WaitForTaskResult wait for task parameters for Codebuild
type WaitForTaskResult struct {
	ID string `json:"id,omitempty"`
}

// StopTaskParams stop task params for Codebuild
type StopTaskParams struct {
	ClusterName string `json:"cluster_name,omitempty" jsonschema:"required"`
	TaskARN     string `json:"task_arn,omitempty" jsonschema:"required"`
}

// StopTaskResult stop task result for Codebuild
type StopTaskResult struct {
	LastStatus string `json:"last_status,omitempty"`
	StopCode   string `json:"stop_reason,omitempty"`

	TaskStatus string `json:"task_status,omitempty"`
}

// CleanupTaskParams cleanup definition params for Codebuild
type CleanupTaskParams struct {
	TaskDefinition string `json:"task_definition,omitempty" jsonschema:"required"`
}

// CleanupTaskResult cleanup definition result for Codebuild
type CleanupTaskResult struct {
}

// GetTaskLogsParams get logs task params for Codebuild
type GetTaskLogsParams struct {
	DefinitionName string `json:"definition_name,omitempty" jsonschema:"required"`
	TaskARN        string `json:"task_arn,omitempty" jsonschema:"required"`

	TaskID    string  `json:"task_id,omitempty" jsonschema:"required"`
	NextToken *string `json:"next_token,omitempty"`
}

// GetTaskLogsResult get logs task result for Codebuild
type GetTaskLogsResult struct {
	LogLines  []*cwlogs.LogLine `json:"log_lines,omitempty"`
	NextToken *string           `json:"next_token,omitempty"`
}
