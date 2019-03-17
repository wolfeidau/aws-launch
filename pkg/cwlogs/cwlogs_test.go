package cwlogs

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wolfeidau/aws-launch/awsmocks"
)

func TestReadLogs(t *testing.T) {

	cwlGetOutput := &cloudwatchlogs.GetLogEventsOutput{
		NextForwardToken: aws.String("f/34139340658027874184690460781927772298499668124394061824"),
		Events: []*cloudwatchlogs.OutputLogEvent{
			&cloudwatchlogs.OutputLogEvent{
				Message: aws.String("test"),
			},
			&cloudwatchlogs.OutputLogEvent{
				Message: aws.String("test"),
			},
		},
	}

	cwlogsSvc := &awsmocks.CloudWatchLogsAPI{}

	cwlogsSvc.On("GetLogEvents", mock.Anything).Return(cwlGetOutput, nil)

	logReader := &CloudwatchLogsReader{cwlogsSvc: cwlogsSvc}

	res, err := logReader.ReadLogs(&ReadLogsParams{})
	require.Nil(t, err)
	require.Len(t, res.LogLines, 2)
	require.Equal(t, "f/34139340658027874184690460781927772298499668124394061824", aws.StringValue(res.NextToken))
}
