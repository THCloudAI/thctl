// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-25
// Description: Aliyun OSS client implementation.

package oss

import (
	"context"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/THCloudAI/thctl/internal/storage"
	"github.com/THCloudAI/thctl/pkg/framework/logger"
)

// Client implements the storage.Provider interface for Aliyun OSS
type Client struct {
	client *oss.Client
	config *storage.Config
	log    *logger.Logger
}

// NewClient creates a new OSS client
func NewClient(config *storage.Config) (*Client, error) {
	log := logger.WithModule("oss")
	
	client, err := oss.New(config.Endpoint, config.AccessKey, config.SecretKey)
	if err != nil {
		log.Errorf("Failed to create OSS client: %v", err)
		return nil, err
	}

	log.Infof("Created OSS client for endpoint %s", config.Endpoint)

	return &Client{
		client: client,
		config: config,
		log:    log,
	}, nil
}

// ListBuckets implements storage.Provider
func (c *Client) ListBuckets(ctx context.Context) ([]storage.Bucket, error) {
	c.log.Debug("Listing buckets")
	
	result, err := c.client.ListBuckets()
	if err != nil {
		c.log.Errorf("Failed to list buckets: %v", err)
		return nil, err
	}

	buckets := make([]storage.Bucket, len(result.Buckets))
	for i, b := range result.Buckets {
		buckets[i] = storage.Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate.String(),
			Region:       b.Location,
		}
	}
	
	c.log.Infof("Listed %d buckets", len(buckets))
	return buckets, nil
}

// ListObjects implements storage.Provider
func (c *Client) ListObjects(ctx context.Context, bucket, prefix string) ([]storage.Object, error) {
	c.log.Debugf("Listing objects in bucket %s with prefix %s", bucket, prefix)
	
	b, err := c.client.Bucket(bucket)
	if err != nil {
		c.log.Errorf("Failed to get bucket %s: %v", bucket, err)
		return nil, err
	}

	lsRes, err := b.ListObjects(oss.Prefix(prefix))
	if err != nil {
		c.log.Errorf("Failed to list objects: %v", err)
		return nil, err
	}

	objects := make([]storage.Object, len(lsRes.Objects))
	for i, obj := range lsRes.Objects {
		objects[i] = storage.Object{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified.String(),
			ETag:         obj.ETag,
		}
	}
	
	c.log.Infof("Listed %d objects", len(objects))
	return objects, nil
}

// UploadObject implements storage.Provider
func (c *Client) UploadObject(ctx context.Context, bucket, key string, reader io.Reader) error {
	c.log.Debugf("Uploading object %s to bucket %s", key, bucket)
	
	b, err := c.client.Bucket(bucket)
	if err != nil {
		c.log.Errorf("Failed to get bucket %s: %v", bucket, err)
		return err
	}

	err = b.PutObject(key, reader)
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
	
	b, err := c.client.Bucket(bucket)
	if err != nil {
		c.log.Errorf("Failed to get bucket %s: %v", bucket, err)
		return err
	}

	body, err := b.GetObject(key)
	if err != nil {
		c.log.Errorf("Failed to get object: %v", err)
		return err
	}
	defer body.Close()

	_, err = io.Copy(writer, body)
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
	
	b, err := c.client.Bucket(bucket)
	if err != nil {
		c.log.Errorf("Failed to get bucket %s: %v", bucket, err)
		return err
	}

	err = b.DeleteObject(key)
	if err != nil {
		c.log.Errorf("Failed to delete object: %v", err)
		return err
	}
	
	c.log.Infof("Successfully deleted object %s", key)
	return nil
}
