package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/wolfeidau/aws-launch/pkg/launcher/codebuild"
	"github.com/wolfeidau/aws-launch/pkg/launcher/ecs"
)

// Dispatcher dispatches definition and launch requests to the correct backend
type Dispatcher struct {
	ECS       ecs.LauncherAPI
	Codebuild codebuild.LauncherAPI
}

// New create a service dispatcher with the AWS configuration overrides
func New(cfgs ...*aws.Config) *Dispatcher {
	return &Dispatcher{
		ECS:       ecs.NewLauncher(cfgs...),
		Codebuild: codebuild.NewLauncher(cfgs...),
	}
}
