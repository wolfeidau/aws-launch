package cwlogs

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CloudwatchLogsReader cloudwatch log reader which uploads chunk of log data to buildkite
type CloudwatchLogsReader struct {
	cwlogsSvc cloudwatchlogsiface.CloudWatchLogsAPI
}

// NewCloudwatchLogsReader read all the things
func NewCloudwatchLogsReader(cfgs ...*aws.Config) *CloudwatchLogsReader {
	sess := session.New(cfgs...)
	return &CloudwatchLogsReader{
		cwlogsSvc: cloudwatchlogs.New(sess),
	}
}

// ReadLogs this reads a page of logs from cloudwatch and returns a token which will access the next page
func (cwlr *CloudwatchLogsReader) ReadLogs(groupName string, streamName string, nextToken string) (string, []byte, error) {

	getlogsInput := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(groupName),
		LogStreamName: aws.String(streamName),
	}

	if nextToken != "" {
		getlogsInput.NextToken = aws.String(nextToken)
	}

	logrus.WithFields(logrus.Fields{
		"LogGroupName":  groupName,
		"LogStreamName": streamName,
		"NextToken":     nextToken,
	}).Info("GetLogEvents")

	getlogsResult, err := cwlr.cwlogsSvc.GetLogEvents(getlogsInput)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to read logs from codebuild cloudwatch log group")
	}

	buf := new(bytes.Buffer)

	for _, event := range getlogsResult.Events {
		_, err := buf.WriteString(aws.StringValue(event.Message))
		if err != nil {
			return "", nil, errors.Wrap(err, "failed to append to buffer")
		}
	}

	nextTokenResult := nextToken

	// only update the token some events came through
	if len(getlogsResult.Events) != 0 {
		nextTokenResult = aws.StringValue(getlogsResult.NextForwardToken)
	}

	return nextTokenResult, buf.Bytes(), nil
}
