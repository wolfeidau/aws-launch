package main

import (
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/onrik/logrus/filename"
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

	newTask             = app.Command("new-task", "Launch a new ECS task.")
	clusterName         = newTask.Arg("cluster-name", "The name of the ecs cluster to run fargate task").Required().String()
	launchContainerName = newTask.Arg("container-name", "The name of the container.").Required().String()
	serviceName         = newTask.Arg("service-name", "The name of the service.").Required().String()
	launchDefName       = newTask.Arg("def-name", "The task definition to launch.").Required().String()
	subnets             = newTask.Arg("subnets", "The vpc subnets to use when launching the container.").Required().String()
	launchCPU           = newTask.Flag("CPU", "How much cpu to provide to the task see https://docs.aws.amazon.com/AmazonECS/latest/developerguide/AWS_Fargate.html").Default("256").Int64()
	launchMemory        = newTask.Flag("memory", "How much memory to provide to the task see https://docs.aws.amazon.com/AmazonECS/latest/developerguide/AWS_Fargate.html").Default("512").Int64()
)

func main() {

	logrus.AddHook(filename.NewHook())
	config := aws.NewConfig()

	lch := launcher.NewECSLauncher(config)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case newDef.FullCommand():
		logrus.WithFields(logrus.Fields{
			"name": *defName,
		}).Info("new definition")

		defTag, err := lch.CreateDefinition(&launcher.DefinitionParams{
			ECS: &launcher.ECSDefinitionParams{
				DefinitionName:   *defName,
				ExecutionRoleARN: *execRoleARN,
				ContainerName:    *containerName,
			},
			Region:      *awsRegion,
			TaskRoleARN: emptyToNil(taskRoleARN),
			Image:       *image,
		})
		if err != nil {
			logrus.Fatalf("failed to create definition: %v", err)
		}

		logrus.WithField("defTag", defTag).Info("created")

	case newTask.FullCommand():
		logrus.WithFields(logrus.Fields{
			"name": *serviceName,
		}).Info("new task")

		err := lch.RunTask(&launcher.RunTaskParams{
			ServiceName:    *serviceName,
			ClusterName:    *clusterName,
			TaskDefinition: *launchDefName,
			ContainerName:  *launchContainerName,
			Environment: map[string]string{
				"RUNNER": "fargate-run-job",
			},
			CPU:     *launchCPU,
			Memory:  *launchMemory,
			Subnets: strings.Split(*subnets, ","),
		})
		if err != nil {
			logrus.WithError(err).Fatal("failed to launch task")
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
