package launcher

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
)

// ServiceDispatcher dispatches definition and launch requests to the correct backend
type ServiceDispatcher struct {
	ECS       Launcher
	Codebuild Launcher
}

// New create a service dispatcher with the AWS configuration overrides
func New(cfgs ...*aws.Config) *ServiceDispatcher {
	return &ServiceDispatcher{
		ECS:       NewECSLauncher(cfgs...),
		Codebuild: NewCodeBuildLauncher(cfgs...),
	}
}

// CreateDefinition create a defintion, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) CreateDefinition(dp *DefinitionParams) (*CreateDefinitionResult, error) {

	if err := dp.Valid(); err != nil {
		return nil, err
	}

	switch {
	case dp.ECS != nil:
		return s.ECS.CreateDefinition(dp)
	case dp.Codebuild != nil:
		return s.Codebuild.CreateDefinition(dp)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}

// RunTask run a task, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) RunTask(rt *RunTaskParams) (*RunTaskResult, error) {
	if err := rt.Valid(); err != nil {
		return nil, err
	}

	switch {
	case rt.ECS != nil:
		return s.ECS.RunTask(rt)
	case rt.Codebuild != nil:
		return s.Codebuild.RunTask(rt)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}
