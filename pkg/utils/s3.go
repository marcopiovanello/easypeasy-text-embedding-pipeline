package utils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3ClientOpts struct {
	Region       string
	Session      string
	Endpoint     string
	AccessKeyId  string
	SecretKeyId  string
	UsePathStyle bool
}

func NewS3Client(opts S3ClientOpts) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(opts.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.AccessKeyId, opts.SecretKeyId, opts.Session),
		),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(opts.Endpoint)
		o.UsePathStyle = opts.UsePathStyle
	})

	return s3Client, nil
}
