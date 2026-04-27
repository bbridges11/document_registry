package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// Backend implements the S3 storage backend
type Backend struct {
	client *s3.Client
	bucket string
}

func NewBackend(client *s3.Client, cfg config.S3Config) *Backend {
	return &Backend{
		client: client,
		bucket: cfg.Bucket,
	}
}

func (b *Backend) Type() shared.StorageBackend {
	return shared.StorageBackendS3
}

func (b *Backend) Upload(ctx context.Context, key string, content io.Reader) (err error) {
	defer err2.Handle(&err)

	try.To1(b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
		Body:   content,
	}))

	return nil
}

func (b *Backend) Download(ctx context.Context, key string) (_ io.ReadCloser, err error) {
	defer err2.Handle(&err)

	result := try.To1(b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}))

	return result.Body, nil
}

func (b *Backend) Delete(ctx context.Context, key string) (err error) {
	defer err2.Handle(&err)

	try.To1(b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}))

	return nil
}

func (b *Backend) Exists(ctx context.Context, key string) (exists bool, err error) {
	defer err2.Handle(&err)

	_, err = b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return false, nil
	}

	return true, nil
}

func (b *Backend) GenerateUploadURL(ctx context.Context, key string) (url string, err error) {
	defer err2.Handle(&err)

	presignClient := s3.NewPresignClient(b.client)

	result := try.To1(presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}))

	return result.URL, nil
}
