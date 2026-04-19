package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type StorageAdapter struct {
	client *s3.Client
	bucket string
}

func NewStorageAdapter(client *s3.Client, cfg config.S3Config) *StorageAdapter {
	return &StorageAdapter{
		client: client,
		bucket: cfg.Bucket,
	}
}

func (a *StorageAdapter) Upload(ctx context.Context, key string, content io.Reader) (err error) {
	defer err2.Handle(&err)

	try.To1(a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
		Body:   content,
	}))

	return nil
}

func (a *StorageAdapter) Download(ctx context.Context, key string) (_ io.ReadCloser, err error) {
	defer err2.Handle(&err)

	result := try.To1(a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	}))

	return result.Body, nil
}

func (a *StorageAdapter) Delete(ctx context.Context, key string) (err error) {
	defer err2.Handle(&err)

	try.To1(a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	}))

	return nil
}

func (a *StorageAdapter) Exists(ctx context.Context, key string) (exists bool, err error) {
	defer err2.Handle(&err)

	_, err = a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return false, nil
	}

	return true, nil
}

func (a *StorageAdapter) GenerateUploadURL(ctx context.Context, key string) (url string, err error) {
	defer err2.Handle(&err)

	presignClient := s3.NewPresignClient(a.client)

	result := try.To1(presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	}))

	return result.URL, nil
}
