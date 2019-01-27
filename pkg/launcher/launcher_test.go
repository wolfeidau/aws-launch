package launcher

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/require"
)

func Test_ShortenARN(t *testing.T) {
	v := shortenTaskArn(aws.String("arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/abcefg1234567890abcefg1234567890"))
	require.Equal(t, "abcefg1234567890abcefg1234567890", v)
}
