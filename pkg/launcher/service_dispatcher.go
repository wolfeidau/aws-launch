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

// DefineAndLaunch create a defintion, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) DefineAndLaunch(dp *DefineAndLaunchParams) (*DefineAndLaunchResult, error) {

	if err := dp.Valid(); err != nil {
		return nil, err
	}

	switch {
	case dp.ECS != nil:
		return s.ECS.DefineAndLaunch(dp)
	case dp.Codebuild != nil:
		return s.Codebuild.DefineAndLaunch(dp)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}

// DefineTask create a defintion, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) DefineTask(dp *DefineTaskParams) (*DefineTaskResult, error) {

	if err := dp.Valid(); err != nil {
		return nil, err
	}

	switch {
	case dp.ECS != nil:
		return s.ECS.DefineTask(dp)
	case dp.Codebuild != nil:
		return s.Codebuild.DefineTask(dp)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}

// LaunchTask run a task, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) LaunchTask(rt *LaunchTaskParams) (*LaunchTaskResult, error) {
	if err := rt.Valid(); err != nil {
		return nil, err
	}

	switch {
	case rt.ECS != nil:
		return s.ECS.LaunchTask(rt)
	case rt.Codebuild != nil:
		return s.Codebuild.LaunchTask(rt)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}

// GetTaskStatus get task status, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) GetTaskStatus(gts *GetTaskStatusParams) (*GetTaskStatusResult, error) {

	switch {
	case gts.ECS != nil:
		return s.ECS.GetTaskStatus(gts)
	case gts.Codebuild != nil:
		return s.Codebuild.GetTaskStatus(gts)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}

// WaitForTask wait for a task to complete, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) WaitForTask(wft *WaitForTaskParams) (*WaitForTaskResult, error) {
	// if err := wft.Valid(); err != nil {
	// 	return nil, err
	// }

	switch {
	case wft.ECS != nil:
		return s.ECS.WaitForTask(wft)
	case wft.Codebuild != nil:
		return s.Codebuild.WaitForTask(wft)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}

// CleanupTask clean up task definition, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) CleanupTask(ctp *CleanupTaskParams) (*CleanupTaskResult, error) {
	if err := ctp.Valid(); err != nil {
		return nil, err
	}

	switch {
	case ctp.ECS != nil:
		return s.ECS.CleanupTask(ctp)
	case ctp.Codebuild != nil:
		return s.Codebuild.CleanupTask(ctp)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}

// GetTaskLogs get the logs for a task, internally this is dispatched to the correct AWS service for creation
func (s *ServiceDispatcher) GetTaskLogs(gtlp *GetTaskLogsParams) (*GetTaskLogsResult, error) {
	if err := gtlp.Valid(); err != nil {
		return nil, err
	}

	switch {
	case gtlp.ECS != nil:
		return s.ECS.GetTaskLogs(gtlp)
	case gtlp.Codebuild != nil:
		return s.Codebuild.GetTaskLogs(gtlp)
	default:
		return nil, errors.New("unable to locate handler for service")
	}
}
