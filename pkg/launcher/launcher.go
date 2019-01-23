package launcher

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/sirupsen/logrus"
)

type Launcher struct {
	ecsSvc    ecsiface.ECSAPI
	cwlogsSvc cloudwatchlogsiface.CloudWatchLogsAPI
}

type LauncherParams struct {
	ClusterName    string
	ServiceName    string
	Environment    map[string]string
	TaskDefinition string
	Subnets        []string
}

type DefinitionParams struct {
	ExecutionRoleARN string
	Region           string
	TaskRoleARN      *string // optional
	DefinitionName   string
	ContainerName    string
	Image            string
}

func New() *Launcher {
	sess := session.New()
	return &Launcher{
		ecsSvc:    ecs.New(sess),
		cwlogsSvc: cloudwatchlogs.New(sess),
	}
}

func (lc *Launcher) CreateDefinition(dp *DefinitionParams) (string, error) {

	_, err := lc.cwlogsSvc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(fmt.Sprintf("/ecs/fargate/%s", dp.DefinitionName)),
		Tags: map[string]*string{
			"createdBy": aws.String("fargate-run-job"),
		},
	})
	if err, ok := err.(awserr.Error); ok {
		if err.Code() != "ResourceAlreadyExistsException" {
			return "", err
		}
	}

	res, err := lc.ecsSvc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		RequiresCompatibilities: aws.StringSlice([]string{
			"FARGATE",
		}),
		Family:      aws.String(dp.DefinitionName),
		TaskRoleArn: dp.TaskRoleARN,
		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
		Cpu:         aws.String("256"),
		Memory:      aws.String("512"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&ecs.ContainerDefinition{
				Name:    aws.String(dp.ContainerName),
				Image:   aws.String(dp.Image),
				Command: aws.StringSlice([]string{"whatever"}),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-group":         aws.String(fmt.Sprintf("/ecs/fargate/%s", dp.DefinitionName)),
						"awslogs-region":        aws.String(dp.Region),
						"awslogs-stream-prefix": aws.String("ecs"),
					},
				},
			},
		},
		ExecutionRoleArn: aws.String(dp.ExecutionRoleARN),
		Tags: []*ecs.Tag{
			&ecs.Tag{
				Key:   aws.String("createdBy"),
				Value: aws.String("fargate-run-job"),
			},
		},
	})
	if err != nil {
		return "", err
	}

	logrus.WithField("result", res).Debug("Register Task Definition")

	return fmt.Sprintf("%s:%d", aws.StringValue(res.TaskDefinition.Family), aws.Int64Value(res.TaskDefinition.Revision)), nil
}

func (lc *Launcher) LauncherTask(lp *LauncherParams) error {

	res, err := lc.ecsSvc.RunTask(&ecs.RunTaskInput{
		Cluster:         aws.String(lp.ClusterName),
		LaunchType:      aws.String(ecs.LaunchTypeFargate),
		TaskDefinition:  aws.String(lp.TaskDefinition),
		Count:           aws.Int64(1),
		PlatformVersion: aws.String("LATEST"),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				Subnets:        aws.StringSlice(lp.Subnets),
			},
		},
		Tags: []*ecs.Tag{
			&ecs.Tag{
				Key:   aws.String("createdBy"),
				Value: aws.String("fargate-run-job"),
			},
		},
	})
	if err != nil {
		return err
	}

	logrus.WithField("result", res).Info("Run Task")

	return nil
}
