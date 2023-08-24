package storage

import (
	"bytes"
	"context"
	"io"
	"time"

	"time-capsule/internal/domain"

	"github.com/minio/minio-go/v7"
)

const timeout = 5 * time.Second

type MinioStorage struct {
	client     *minio.Client
	bucketName string
}

func NewMinioStorage(client *minio.Client, bucketName string) Storage {
	return &MinioStorage{
		client:     client,
		bucketName: bucketName,
	}
}

func (s *MinioStorage) Delete(ctx context.Context, fileName string) error {
	opts := minio.RemoveObjectOptions{}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.client.RemoveObject(
		ctx,
		s.bucketName,
		fileName,
		opts,
	)
}

func (s *MinioStorage) Get(ctx context.Context, fileName string) (*domain.File, error) {
	opts := minio.GetObjectOptions{}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	obj, err := s.client.GetObject(ctx, s.bucketName, fileName, opts)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	objInfo, err := obj.Stat()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, objInfo.Size)
	_, err = obj.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	file := &domain.File{
		Name:  fileName,
		Size:  objInfo.Size,
		Bytes: buffer,
	}

	return file, nil
}

func (s *MinioStorage) Upload(ctx context.Context, file domain.File) error {
	opts := minio.PutObjectOptions{
		ContentType: "image/png",
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	reader := bytes.NewReader(file.Bytes)

	_, err := s.client.PutObject(
		ctx,
		s.bucketName,
		file.Name,
		reader,
		file.Size,
		opts,
	)

	return err
}
