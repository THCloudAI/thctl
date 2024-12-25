// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-25
// Description: Common storage types and interfaces.

package storage

import (
	"context"
	"io"
)

// Provider represents a storage provider interface
type Provider interface {
	// ListBuckets lists all buckets
	ListBuckets(ctx context.Context) ([]Bucket, error)
	// ListObjects lists objects in a bucket with optional prefix
	ListObjects(ctx context.Context, bucket, prefix string) ([]Object, error)
	// UploadObject uploads an object to the storage
	UploadObject(ctx context.Context, bucket, key string, reader io.Reader) error
	// DownloadObject downloads an object from the storage
	DownloadObject(ctx context.Context, bucket, key string, writer io.Writer) error
	// DeleteObject deletes an object from the storage
	DeleteObject(ctx context.Context, bucket, key string) error
}

// Bucket represents a storage bucket
type Bucket struct {
	Name         string
	CreationDate string
	Region       string
}

// Object represents a storage object
type Object struct {
	Key          string
	Size         int64
	LastModified string
	ETag         string
}

// Config represents common storage configuration
type Config struct {
	Region     string `mapstructure:"region"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	Endpoint   string `mapstructure:"endpoint"`
	BucketName string `mapstructure:"bucket_name"`
}
