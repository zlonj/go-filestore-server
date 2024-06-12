package s3

import (
	"context"
	"fmt"

	appConfig "filestore-server/config"

	aswConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var client *s3.Client

func Client() *s3.Client {
	if client != nil {
		return client
	}

	config, err := aswConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	client = s3.NewFromConfig(config, func(o *s3.Options) {
		o.Region = appConfig.S3_REGION
	})
	return client
}

