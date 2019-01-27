package launcher

import (
	"github.com/pkg/errors"
)

var (
	// ErrMissingECSParams missing the params required by the ecs launch
	ErrMissingECSParams = errors.New("ECS params are missing from Definition")
)

// Launcher build the definition, then launch a container based task
type Launcher interface {
	CreateDefinition(*DefinitionParams) (string, error)
	RunTask(*RunTaskParams) error
}

// RunTaskParams used to launch container based tasks
type RunTaskParams struct {
	ECS         *ECSTaskParams    `json:"ecs,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Subnets     []string          `json:"subnets,omitempty" jsonschema:"required"`
	CPU         int64             `json:"cpu,omitempty" jsonschema:"required"`
	Memory      int64             `json:"memory,omitempty" jsonschema:"required"`
}

// ECSTaskParams ECS related run task parameters
type ECSTaskParams struct {
	ClusterName    string `json:"cluster_name,omitempty" jsonschema:"required"`
	ServiceName    string `json:"service_name,omitempty" jsonschema:"required"`
	ContainerName  string `json:"container_name,omitempty" jsonschema:"required"`
	TaskDefinition string `json:"task_definition,omitempty" jsonschema:"required"`
}

// ECSDefinitionParams ECS related definition parameters
type ECSDefinitionParams struct {
	ExecutionRoleARN string `json:"execution_role_arn,omitempty" jsonschema:"required"`
	DefinitionName   string `json:"definition_name,omitempty" jsonschema:"required"`
	ContainerName    string `json:"container_name,omitempty" jsonschema:"required"`
}

// DefinitionParams parameters used to build a container execution environment
type DefinitionParams struct {
	ECS         *ECSDefinitionParams `json:"ecs,omitempty"`
	Region      string               `json:"region,omitempty" jsonschema:"required"`
	TaskRoleARN *string              `json:"task_role_arn,omitempty"` // optional
	Image       string               `json:"image,omitempty" jsonschema:"required"`
	Environment map[string]string    `json:"environment,omitempty"`
	Tags        map[string]string    `json:"tags,omitempty"`
}
