package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	guitarKeyPrefix    = "images/guitars"
	marketLogKeyPrefix = "images/market-logs"
)

var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// PresignResult holds the URLs returned to the client for a single upload.
type PresignResult struct {
	UploadURL string
	PublicURL string
	Key       string
}

// Presigner generates S3 presigned PUT URLs for guitar pictures.
type Presigner struct {
	presign         *s3.PresignClient
	bucket          string
	cdnBase         string
	publicS3BaseURL string
}

// NewPresigner constructs a Presigner. bucket is the S3 bucket name; cdnBaseURL
// is the public origin used to serve uploaded objects (CloudFront or S3 URL).
// publicS3BaseURL, when set, rewrites presigned PUT URLs so browsers can reach
// S3 (e.g. http://localhost:4566 in LocalStack while Lambda uses the docker hostname).
func NewPresigner(client *s3.Client, bucket, cdnBaseURL, publicS3BaseURL string) *Presigner {
	return &Presigner{
		presign:         s3.NewPresignClient(client),
		bucket:          bucket,
		cdnBase:         strings.TrimRight(cdnBaseURL, "/"),
		publicS3BaseURL: strings.TrimRight(publicS3BaseURL, "/"),
	}
}

// PresignPut validates contentType and returns a short-lived PUT URL plus the
// stable public URL clients should store on the guitar record.
func (p *Presigner) PresignPut(ctx context.Context, contentType string) (*PresignResult, error) {
	return p.presignPut(ctx, guitarKeyPrefix, contentType)
}

// PresignMarketLogImage returns a presigned PUT URL for a crawled listing thumbnail.
func (p *Presigner) PresignMarketLogImage(ctx context.Context, contentType string) (*PresignResult, error) {
	return p.presignPut(ctx, marketLogKeyPrefix, contentType)
}

func (p *Presigner) presignPut(ctx context.Context, prefix, contentType string) (*PresignResult, error) {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	ext, ok := allowedContentTypes[contentType]
	if !ok {
		return nil, fmt.Errorf("unsupported content type %q", contentType)
	}

	key := fmt.Sprintf("%s/%s%s", prefix, uuid.NewString(), ext)

	out, err := p.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return nil, err
	}

	uploadURL := out.URL
	if p.publicS3BaseURL != "" {
		uploadURL, err = rewritePresignedHost(out.URL, p.publicS3BaseURL)
		if err != nil {
			return nil, err
		}
	}

	return &PresignResult{
		UploadURL: uploadURL,
		PublicURL: fmt.Sprintf("%s/%s", p.cdnBase, key),
		Key:       key,
	}, nil
}

func rewritePresignedHost(rawURL, publicBase string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	pub, err := url.Parse(publicBase)
	if err != nil {
		return "", err
	}
	u.Scheme = pub.Scheme
	u.Host = pub.Host
	return u.String(), nil
}
