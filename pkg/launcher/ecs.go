package launcher

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultCPU default cpu allocation to a fargate task
	DefaultCPU = "256"
	// DefaultMemory default memory allocation to a fargate task
	DefaultMemory = "512"
)

// ECSLauncher used to launch containers in ECS, specifically fargate
type ECSLauncher struct {
	ecsSvc    ecsiface.ECSAPI
	cwlogsSvc cloudwatchlogsiface.CloudWatchLogsAPI
}

// NewECSLauncher create a new launcher
func NewECSLauncher(cfgs ...*aws.Config) *ECSLauncher {
	sess := session.New(cfgs...)
	return &ECSLauncher{
		ecsSvc:    ecs.New(sess),
		cwlogsSvc: cloudwatchlogs.New(sess),
	}
}

// CreateDefinition create a container task definition
func (lc *ECSLauncher) CreateDefinition(dp *DefinitionParams) (string, error) {

	_, err := lc.cwlogsSvc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(fmt.Sprintf("/ecs/fargate/%s", dp.ECS.DefinitionName)),
		Tags: map[string]*string{
			"createdBy": aws.String("fargate-run-job"),
		},
	})
	if err, ok := err.(awserr.Error); ok {
		if err.Code() != "ResourceAlreadyExistsException" {
			return "", errors.Wrap(err, "create log group failed.")
		}
	}

	// register the task definition with default base memory, cpu and cwlogs groups
	res, err := lc.ecsSvc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		RequiresCompatibilities: aws.StringSlice([]string{
			"FARGATE",
		}),
		Family:      aws.String(dp.ECS.DefinitionName),
		TaskRoleArn: dp.TaskRoleARN,
		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
		Cpu:         aws.String(DefaultCPU),
		Memory:      aws.String(DefaultMemory),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&ecs.ContainerDefinition{
				Name:  aws.String(dp.ECS.ContainerName),
				Image: aws.String(dp.Image),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-group":         aws.String(fmt.Sprintf("/ecs/fargate/%s", dp.ECS.DefinitionName)),
						"awslogs-region":        aws.String(dp.Region),
						"awslogs-stream-prefix": aws.String("ecs"),
					},
				},
				Environment: convertMapToKeyValuePair(dp.Environment),
			},
		},
		ExecutionRoleArn: aws.String(dp.ECS.ExecutionRoleARN),
		Tags:             convertMapToECSTags(dp.Tags),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to register task definition.")
	}

	logrus.WithField("result", res).Debug("Register Task Definition")

	return fmt.Sprintf("%s:%d", aws.StringValue(res.TaskDefinition.Family), aws.Int64Value(res.TaskDefinition.Revision)), nil
}

// RunTask run a container task and monitor it till completion
func (lc *ECSLauncher) RunTask(lp *RunTaskParams) error {

	logrus.WithFields(logrus.Fields{
		"ClusterName":    lp.ECS.ClusterName,
		"TaskDefinition": lp.ECS.TaskDefinition,
	}).Info("Launch Task")

	taskRes, err := lc.ecsSvc.RunTask(&ecs.RunTaskInput{
		Cluster:        aws.String(lp.ECS.ClusterName),
		LaunchType:     aws.String(ecs.LaunchTypeFargate),
		TaskDefinition: aws.String(lp.ECS.TaskDefinition),
		Count:          aws.Int64(1),
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				&ecs.ContainerOverride{
					Cpu:         aws.Int64(lp.ECS.CPU),
					Memory:      aws.Int64(lp.ECS.Memory),
					Name:        aws.String(lp.ECS.ContainerName),
					Environment: convertMapToKeyValuePair(lp.Environment),
				},
			},
		},
		PlatformVersion: aws.String("LATEST"),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				Subnets:        aws.StringSlice(lp.ECS.Subnets),
			},
		},
		Tags: convertMapToECSTags(lp.Tags),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create task.")
	}

	logrus.WithFields(logrus.Fields{
		"TaskID": shortenTaskArn(taskRes.Tasks[0].TaskArn),
	}).Info("Task Provisioned")

	descInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(lp.ECS.ClusterName),
		Tasks:   []*string{taskRes.Tasks[0].TaskArn},
	}

	err = lc.ecsSvc.WaitUntilTasksStopped(descInput)
	if err != nil {
		return errors.Wrap(err, "failed to check stopped task.")
	}

	descRes, err := lc.ecsSvc.DescribeTasks(descInput)
	if err != nil {
		return errors.Wrap(err, "failed to describe task.")
	}

	logrus.WithFields(logrus.Fields{
		"TaskID":        shortenTaskArn(descRes.Tasks[0].TaskArn),
		"StopCode":      aws.StringValue(descRes.Tasks[0].StopCode),
		"StoppedReason": aws.StringValue(descRes.Tasks[0].StoppedReason),
	}).Info("Describe completed Task")

	return nil
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

	// empty map is valid
	if env == nil {
		return nil
	}

	for k, v := range env {
		ecsEnv = append(ecsEnv, &ecs.KeyValuePair{Name: aws.String(k), Value: aws.String(v)})
	}

	return ecsEnv
}

func convertMapToECSTags(tags map[string]string) []*ecs.Tag {

	ecsTags := []*ecs.Tag{}

	// empty map is valid
	if tags == nil {
		return nil
	}

	for k, v := range tags {
		ecsTags = append(ecsTags, &ecs.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	return ecsTags
}
