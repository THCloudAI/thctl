// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-25
// Description: Tencent COS client implementation.

package cos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/THCloudAI/thctl/internal/storage"
	"github.com/THCloudAI/thctl/pkg/framework/logger"
)

// Client implements the storage.Provider interface for Tencent COS
type Client struct {
	client *cos.Client
	config *storage.Config
	log    *logger.Logger
}

// NewClient creates a new COS client
func NewClient(config *storage.Config) (*Client, error) {
	log := logger.WithModule("cos")
	
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.BucketName, config.Region))
	if err != nil {
		log.Errorf("Failed to parse COS URL: %v", err)
		return nil, err
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.AccessKey,
			SecretKey: config.SecretKey,
		},
	})

	log.Infof("Created COS client for bucket %s in region %s", config.BucketName, config.Region)

	return &Client{
		client: client,
		config: config,
		log:    log,
	}, nil
}

// ListBuckets implements storage.Provider
func (c *Client) ListBuckets(ctx context.Context) ([]storage.Bucket, error) {
	c.log.Debug("Listing buckets")
	
	result, _, err := c.client.Service.Get(ctx)
	if err != nil {
		c.log.Errorf("Failed to list buckets: %v", err)
		return nil, err
	}

	buckets := make([]storage.Bucket, len(result.Buckets))
	for i, b := range result.Buckets {
		buckets[i] = storage.Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate,
			Region:       b.Region,
		}
	}
	
	c.log.Infof("Listed %d buckets", len(buckets))
	return buckets, nil
}

// ListObjects implements storage.Provider
func (c *Client) ListObjects(ctx context.Context, bucket, prefix string) ([]storage.Object, error) {
	c.log.Debugf("Listing objects in bucket %s with prefix %s", bucket, prefix)
	
	opt := &cos.BucketGetOptions{
		Prefix: prefix,
	}
	result, _, err := c.client.Bucket.Get(ctx, opt)
	if err != nil {
		c.log.Errorf("Failed to list objects: %v", err)
		return nil, err
	}

	objects := make([]storage.Object, len(result.Contents))
	for i, obj := range result.Contents {
		objects[i] = storage.Object{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ETag:         obj.ETag,
		}
	}
	
	c.log.Infof("Listed %d objects", len(objects))
	return objects, nil
}

// UploadObject implements storage.Provider
func (c *Client) UploadObject(ctx context.Context, bucket, key string, reader io.Reader) error {
	c.log.Debugf("Uploading object %s to bucket %s", key, bucket)
	
	_, err := c.client.Object.Put(ctx, key, reader, nil)
	if err != nil {
		c.log.Errorf("Failed to upload object: %v", err)
		return err
	}
	
	c.log.Infof("Successfully uploaded object %s", key)
	return nil
}

// DownloadObject implements storage.Provider
func (c *Client) DownloadObject(ctx context.Context, bucket, key string, writer io.Writer) error {
	c.log.Debugf("Downloading object %s from bucket %s", key, bucket)
	
	resp, err := c.client.Object.Get(ctx, key, nil)
	if err != nil {
		c.log.Errorf("Failed to download object: %v", err)
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		c.log.Errorf("Failed to write object data: %v", err)
		return err
	}
	
	c.log.Infof("Successfully downloaded object %s", key)
	return nil
}

// DeleteObject implements storage.Provider
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	c.log.Debugf("Deleting object %s from bucket %s", key, bucket)
	
	_, err := c.client.Object.Delete(ctx, key)
	if err != nil {
		c.log.Errorf("Failed to delete object: %v", err)
		return err
	}
	
	c.log.Infof("Successfully deleted object %s", key)
	return nil
}
