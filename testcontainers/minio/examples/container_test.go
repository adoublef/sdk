// https://github.com/minio/minio/blob/master/docs/orchestration/docker-compose/docker-compose.yaml
// https://github.com/mmadfox/testcontainers/blob/master/minio/minio_test.go
// https://github.com/mmadfox/testcontainers/blob/master/minio/minio.go
// https://github.com/testcontainers/testcontainers-go/blob/main/examples/cockroachdb/cockroachdb.go
// https://dev.to/minhblues/easy-file-uploads-in-go-fiber-with-minio-393c
package examples

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/adoublef/sdk/bytest"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMinio(t *testing.T) {
	ctx := context.Background()

	minioC, err := setupMinio(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := minioC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	c, err := minio.New(minioC.URI, &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""), // seems to play no affect
		Secure: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	bucketName := "testcontainers"
	location := "eu-west-2"

	// create bucket
	err = c.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		t.Fatal(err)
	}

	objectName := "testdata"
	contentType := "applcation/octet-stream"

	uploadInfo, err := c.PutObject(ctx, bucketName, objectName, bytest.NewReader(bytest.MB*16), (bytest.MB * 16), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		t.Fatal(err)
	}

	// object is a readSeekCloser
	object, err := c.GetObject(ctx, uploadInfo.Bucket, uploadInfo.Key, minio.GetObjectOptions{})
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

type minioContainer struct {
	testcontainers.Container
	URI string
}

func setupMinio(ctx context.Context) (*minioContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:RELEASE.2024-01-16T16-07-38Z",
		ExposedPorts: []string{"9000/tcp", "9001/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		},
		Cmd:        []string{"server", "/data"},
		WaitingFor: wait.ForListeningPort("9000").WithStartupTimeout(time.Minute * 2),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "9000")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	return &minioContainer{Container: container, URI: uri}, nil
}
