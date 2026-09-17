package storage

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Exercises the driver against a real S3-compatible server, which is the only
// way to know the signing, path-style addressing and checksum settings are
// right -- a mock would just agree with whatever the code does.
//
// Skipped unless TEST_S3_ENDPOINT is set, so `go test ./...` needs no
// infrastructure. Run MinIO with:
//
//	docker run -d -p 9100:9000 -e MINIO_ROOT_USER=testkey \
//	  -e MINIO_ROOT_PASSWORD=testsecret123 quay.io/minio/minio server /data
func testConfig(t *testing.T) S3Config {
	t.Helper()
	endpoint := os.Getenv("TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("TEST_S3_ENDPOINT not set; skipping object storage test")
	}
	return S3Config{
		Endpoint:  endpoint,
		Region:    envOr("TEST_S3_REGION", "us-east-1"),
		Bucket:    envOr("TEST_S3_BUCKET", "aperture-test"),
		AccessKey: envOr("TEST_S3_ACCESS_KEY", "testkey"),
		SecretKey: envOr("TEST_S3_SECRET_KEY", "testsecret123"),
		Prefix:    "generations",
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func ensureBucket(t *testing.T, cfg S3Config) {
	t.Helper()
	client := s3.NewFromConfig(aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey, cfg.SecretKey, ""),
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
	}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})
	_, err := client.CreateBucket(context.Background(),
		&s3.CreateBucketInput{Bucket: aws.String(cfg.Bucket)})
	if err != nil && !strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") &&
		!strings.Contains(err.Error(), "BucketAlreadyExists") {
		t.Fatalf("create bucket: %v", err)
	}
}

func TestS3PutStoresAndReturnsAFetchableURL(t *testing.T) {
	cfg := testConfig(t)
	ensureBucket(t, cfg)

	store, err := NewS3(cfg)
	if err != nil {
		t.Fatal(err)
	}

	body := []byte("\x89PNG\r\n\x1a\n-- pretend this is an image --")
	url, err := store.Put(context.Background(), "result.png", body, "image/png")
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if !strings.Contains(url, cfg.Bucket) || !strings.HasSuffix(url, "result.png") {
		t.Fatalf("unexpected url: %s", url)
	}
	if !strings.Contains(url, "generations/") {
		t.Fatalf("prefix missing from key: %s", url)
	}

	// Read it back through the S3 API rather than the public URL, so the test
	// does not depend on the bucket being anonymously readable.
	client := s3.NewFromConfig(aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey, cfg.SecretKey, ""),
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
	}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	key := strings.SplitN(strings.TrimPrefix(url, cfg.Endpoint+"/"), "/", 2)
	if len(key) != 2 {
		t.Fatalf("could not derive key from %s", url)
	}
	obj, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket), Key: aws.String(key[1]),
	})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer obj.Body.Close()

	got, _ := io.ReadAll(obj.Body)
	if string(got) != string(body) {
		t.Fatalf("content round-trip failed: %d bytes back, %d sent", len(got), len(body))
	}
	if obj.ContentType == nil || *obj.ContentType != "image/png" {
		t.Fatalf("content type lost: %v", obj.ContentType)
	}
}

// Two uploads of the same filename must not collide, or one generation would
// overwrite another's result.
func TestS3KeysAreUnique(t *testing.T) {
	cfg := testConfig(t)
	ensureBucket(t, cfg)
	store, _ := NewS3(cfg)

	a, err := store.Put(context.Background(), "same.png", []byte("a"), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	b, err := store.Put(context.Background(), "same.png", []byte("b"), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatalf("two uploads produced the same key: %s", a)
	}
}

func TestNewS3RejectsIncompleteConfig(t *testing.T) {
	for name, cfg := range map[string]S3Config{
		"no bucket":   {Endpoint: "https://x", AccessKey: "a", SecretKey: "b"},
		"no endpoint": {Bucket: "b", AccessKey: "a", SecretKey: "b"},
		"no creds":    {Bucket: "b", Endpoint: "https://x"},
	} {
		if _, err := NewS3(cfg); err == nil {
			t.Fatalf("%s: accepted an incomplete config", name)
		}
	}
}
