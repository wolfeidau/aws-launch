package launcher

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
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
	ClusterName    string            `json:"cluster_name,omitempty"`
	ServiceName    string            `json:"service_name,omitempty"`
	ContainerName  string            `json:"container_name,omitempty"`
	Environment    map[string]string `json:"environment,omitempty"`
	TaskDefinition string            `json:"task_definition,omitempty"`
	Subnets        []string          `json:"subnets,omitempty"`
	CPU            int64             `json:"cpu,omitempty"`
	Memory         int64             `json:"memory,omitempty"`
}

// ECSDefinitionParams ECS related definition parameters
type ECSDefinitionParams struct {
	ExecutionRoleARN string `json:"execution_role_arn,omitempty"`
	DefinitionName   string `json:"definition_name,omitempty"`
	ContainerName    string `json:"container_name,omitempty"`
}

// DefinitionParams parameters used to build a container execution environment
type DefinitionParams struct {
	ECS         *ECSDefinitionParams `json:"ecs,omitempty"`
	Region      string               `json:"region,omitempty"`
	TaskRoleARN *string              `json:"task_role_arn,omitempty"` // optional
	Image       string               `json:"image,omitempty"`
}

func shortenTaskArn(taskArn *string) string {
	tokens := strings.Split(aws.StringValue(taskArn), "/")
	if len(tokens) == 3 {
		return tokens[2]
	}

	return "unknown"
}

func convertMapToKeyValuePair(env map[string]string) []*ecs.KeyValuePair {

	ecsEnv := []*ecs.KeyValuePair{}

	for k, v := range env {
		//ecsEnv[]
		ecsEnv = append(ecsEnv, &ecs.KeyValuePair{Name: aws.String(k), Value: aws.String(v)})
	}

	return ecsEnv
}
