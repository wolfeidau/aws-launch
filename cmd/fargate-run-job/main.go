package main

import (
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/fargate-run-job/pkg/launcher"
)

var (
	app       = kingpin.New("fargate-run-job", "A command-line fargate provisioning application.")
	verbose   = app.Flag("verbose", "Verbose mode.").Short('v').Bool()
	awsRegion = app.Flag("aws-region", "The aws region used when creating resources.").Required().String()

	newDef = app.Command("new-definition", "Build a new definition for an ECS task.")

	image         = newDef.Arg("image", "The docker image to use for the task.").Required().String()
	containerName = newDef.Arg("container-name", "The name of the container.").Required().String()
	defName       = newDef.Arg("def-name", "The name of the task definition.").Required().String()
	execRoleARN   = newDef.Arg("exec-arn", "The ARN of the role used for the execution of the container.").Required().String()
	taskRoleARN   = newDef.Flag("task-arn", "The ARN of the role used for the task.").String()

	newTask       = app.Command("new-task", "Launch a new ECS task.")
	clusterName   = newTask.Arg("cluster-name", "The name of the ecs cluster to run fargate task").Required().String()
	serviceName   = newTask.Arg("service-name", "The name of the service.").Required().String()
	launchDefName = newTask.Arg("def-name", "The task definition to launch.").Required().String()
	subnets       = newTask.Arg("subnets", "The vpc subnets to use when launching the container.").Required().String()
)

func main() {
	lch := launcher.New()

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case newDef.FullCommand():
		logrus.WithFields(logrus.Fields{
			"name": *defName,
		}).Info("new definition")

		defTag, err := lch.CreateDefinition(&launcher.DefinitionParams{
			DefinitionName:   *defName,
			Region:           *awsRegion,
			TaskRoleARN:      emptyToNil(taskRoleARN),
			ExecutionRoleARN: *execRoleARN,
			ContainerName:    *containerName,
			Image:            *image,
		})
		if err != nil {
			logrus.Fatalf("failed to create definition: %v", err)
		}

		logrus.WithField("defTag", defTag).Info("created")

	case newTask.FullCommand():
		logrus.WithFields(logrus.Fields{
			"name": *serviceName,
		}).Info("new task")

		err := lch.LauncherTask(&launcher.LauncherParams{
			ServiceName:    *serviceName,
			ClusterName:    *clusterName,
			TaskDefinition: *launchDefName,
			Subnets:        strings.Split(*subnets, ","),
		})
		if err != nil {
			logrus.Fatalf("failed to launch task: %v", err)
		}
	}
}

func emptyToNil(val *string) *string {
	if val == nil {
		return val
	}
	if *val == "" {
		return nil
	}
	return val
}
