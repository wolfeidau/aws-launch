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
	ClusterName    string
	ServiceName    string
	ContainerName  string
	Environment    map[string]string
	TaskDefinition string
	Subnets        []string
	CPU            int64
	Memory         int64
}

// ECSDefinitionParams ECS related definition parameters
type ECSDefinitionParams struct {
	ExecutionRoleARN string
	DefinitionName   string
	ContainerName    string
}

// DefinitionParams parameters used to build a container execution environment
type DefinitionParams struct {
	ECS         *ECSDefinitionParams
	Region      string
	TaskRoleARN *string // optional
	Image       string
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
