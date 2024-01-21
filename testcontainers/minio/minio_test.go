package minio_test

import (
	"context"
	"io"
	"testing"

	"github.com/adoublef/sdk/bytest"
	. "github.com/adoublef/sdk/testcontainers/minio"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestContainer(t *testing.T) {
	ctx := context.Background()

	minioContainer, err := RunContainer(ctx, WithUsername("username"), WithPassword("password"))
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := minioContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	url, err := minioContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	minioClient, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(minioContainer.Username, minioContainer.Password, ""), // seems to play no affect
		Secure: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	bucketName := "testcontainers"
	location := "eu-west-2"

	// create bucket
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		t.Fatal(err)
	}

	objectName := "testdata"
	contentType := "applcation/octet-stream"

	uploadInfo, err := minioClient.PutObject(ctx, bucketName, objectName, bytest.NewReader(bytest.MB*16), (bytest.MB * 16), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		t.Fatal(err)
	}

	// object is a readSeekCloser
	object, err := minioClient.GetObject(ctx, uploadInfo.Bucket, uploadInfo.Key, minio.GetObjectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer object.Close()

	n, err := io.Copy(io.Discard, object)
	if err != nil {
		t.Fatal(err)
	}

	if n != bytest.MB*16 {
		t.Fatalf("expected %d; got %d", bytest.MB*16, n)
	}
}
