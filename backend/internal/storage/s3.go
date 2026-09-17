package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config describes any S3-compatible object store. The names mirror the AWS
// environment variables, so a provider's own quickstart maps onto it directly.
type S3Config struct {
	Endpoint  string // https://br-<branch>.storage.c-2.us-east-2.aws.neon.tech
	Region    string // us-east-2
	Bucket    string
	AccessKey string
	SecretKey string
	// PublicBaseURL serves objects when the bucket allows anonymous reads.
	// Empty means fall back to <endpoint>/<bucket>.
	PublicBaseURL string
	Prefix        string
}

// S3 stores generated media in an S3-compatible bucket.
//
// Written against the S3 API rather than one vendor's SDK, so Neon Object
// Storage, Cloudflare R2, AWS S3 and MinIO are all a matter of configuration.
type S3 struct {
	client    *s3.Client
	bucket    string
	prefix    string
	publicURL string
}

func NewS3(cfg S3Config) (*S3, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("storage: bucket is required")
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("storage: endpoint is required")
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("storage: access key and secret are required")
	}
	if _, err := url.Parse(cfg.Endpoint); err != nil {
		return nil, fmt.Errorf("storage: endpoint is not a valid url: %w", err)
	}

	awsCfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey, cfg.SecretKey, ""),
		// Recent SDKs attach a CRC32 checksum to every request by default.
		// Several S3-compatible stores, Neon included, reject that, so
		// checksums are only sent when the operation actually requires one.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		// Path-style addressing: bucket in the path, not the hostname. Neon
		// requires it, and every other S3-compatible store accepts it.
		o.UsePathStyle = true
	})

	public := cfg.PublicBaseURL
	if public == "" {
		public = strings.TrimSuffix(cfg.Endpoint, "/") + "/" + cfg.Bucket
	}

	return &S3{
		client:    client,
		bucket:    cfg.Bucket,
		prefix:    strings.Trim(cfg.Prefix, "/"),
		publicURL: strings.TrimSuffix(public, "/"),
	}, nil
}

func (s *S3) Put(ctx context.Context, name string, data []byte, contentType string) (string, error) {
	key, err := s.key(name)
	if err != nil {
		return "", err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Long cache lifetime: keys carry random bytes, so an object is never
	// replaced in place and its content can never go stale.
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(key),
		Body:         bytes.NewReader(data),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return "", fmt.Errorf("storage: upload %s: %w", key, err)
	}

	return s.publicURL + "/" + key, nil
}

// key namespaces objects by date and prefixes a random component, so two
// generations can never collide and a listing stays browsable.
func (s *S3) key(name string) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	parts := []string{}
	if s.prefix != "" {
		parts = append(parts, s.prefix)
	}
	parts = append(parts,
		time.Now().UTC().Format("2006/01/02"),
		hex.EncodeToString(buf)+"-"+filepath.Base(name))
	return strings.Join(parts, "/"), nil
}
