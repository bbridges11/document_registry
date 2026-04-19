package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appConfig "github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func NewS3Client(ctx context.Context, cfg appConfig.AWSConfig) (client *s3.Client, err error) {
	defer err2.Handle(&err)

	awsCfg := try.To1(config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
	))

	opts := []func(*s3.Options){}

	// Use custom endpoint for local development
	if cfg.S3.Endpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.S3.Endpoint)
			o.UsePathStyle = true
		})
	}

	return s3.NewFromConfig(awsCfg, opts...), nil
}
