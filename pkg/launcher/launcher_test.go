package launcher

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/service/codebuild"
)

func TestDefineAndLaunchParams_BuildDefineTask_Codebuild(t *testing.T) {

	cb := &CodebuildDefineAndLaunchParams{
		ProjectName: "test123",
		ComputeType: codebuild.ComputeTypeBuildGeneral1Large,
		ServiceRole: "abc123",
		Buildspec:   "version: 0.2\n\nphases:\n\tbuild:\n\t\tcommands:\n\t\t\t- command",
	}

	expected := &DefineTaskParams{
		Region:      "ap-southeast-2",
		Image:       "wolfeidau/test",
		Environment: map[string]string{"NODE_ENV": "testing"},
		Tags:        map[string]string{"RUNTIME": "node"},
		Codebuild: &CodebuildDefineTaskParams{
			ProjectName: "test123",
			ComputeType: codebuild.ComputeTypeBuildGeneral1Large,
			Buildspec:   "version: 0.2\n\nphases:\n\tbuild:\n\t\tcommands:\n\t\t\t- command",
			ServiceRole: "abc123",
		},
	}

	dlp := &DefineAndLaunchParams{
		Codebuild: cb,
		Region:    "ap-southeast-2",
		Image:     "wolfeidau/test",
		Environment: map[string]string{
			"NODE_ENV": "testing",
		},
		Tags: map[string]string{
			"RUNTIME": "node",
		},
	}
	got := dlp.BuildDefineTask()
	require.Equal(t, expected, got)
}
