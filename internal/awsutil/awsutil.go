package awsutil

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"os"
)

//GetDefaultConfig returns an aws config with region filled in
//and localEndpoint set if localEndpoint != ""
func GetDefaultConfig(localEndpoint string) *aws.Config {
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = endpoints.UsEast1RegionID
	}
	awsCfg := aws.NewConfig().WithRegion(region)
	if localEndpoint != "" {
		awsCfg.Endpoint = aws.String(localEndpoint)
	}

	return awsCfg
}
