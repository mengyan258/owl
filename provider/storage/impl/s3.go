package impl

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bit-labs.cn/owl/provider/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Storage S3 存储实现
type S3Storage struct {
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	config     *storage.S3Config
}

// NewS3Storage 创建 S3 存储实例
func NewS3Storage(config *storage.S3Config) (*S3Storage, error) {
	// 创建 AWS 会话
	awsConfig := &aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AccessKeyID,
			config.SecretAccessKey,
			"",
		),
	}

	// 如果配置了自定义端点
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(config.PathStyle)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// 创建 S3 客户端
	client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	return &S3Storage{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		config:     config,
	}, nil
}

// Put 上传文件
func (s *S3Storage) Put(ctx context.Context, path string, reader io.Reader, size int64) (*storage.FileInfo, error) {
	key := s.buildPath(path)

	// 读取数据并计算 MD5
	buf := new(bytes.Buffer)
	hash := md5.New()
	tee := io.TeeReader(reader, hash)

	_, err := buf.ReadFrom(tee)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// 上传文件
	result, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(storage.MimeType(path)),
		ACL:         aws.String(s.getACL()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 构建文件信息
	fileInfo := &storage.FileInfo{
		Name:        filepath.Base(path),
		Path:        path,
		Size:        int64(buf.Len()),
		ContentType: storage.MimeType(path),
		Extension:   filepath.Ext(path),
		URL:         s.buildURL(key),
		Hash:        fmt.Sprintf("%x", hash.Sum(nil)),
		UploadTime:  time.Now(),
		Metadata: map[string]string{
			"bucket":   s.config.Bucket,
			"key":      key,
			"etag":     strings.Trim(*result.ETag, "\""),
			"location": result.Location,
		},
	}

	return fileInfo, nil
}

// PutFile 上传本地文件
func (s *S3Storage) PutFile(ctx context.Context, path string, localPath string) (*storage.FileInfo, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %w", err)
	}

	return s.Put(ctx, path, file, stat.Size())
}

// Get 获取文件
func (s *S3Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	key := s.buildPath(path)

	result, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return result.Body, nil
}

// Delete 删除文件
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	key := s.buildPath(path)

	_, err := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	key := s.buildPath(path)

	_, err := s.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchKey:
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// Size 获取文件大小
func (s *S3Storage) Size(ctx context.Context, path string) (int64, error) {
	key := s.buildPath(path)

	result, err := s.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return *result.ContentLength, nil
}

// URL 获取文件访问 URL
func (s *S3Storage) URL(ctx context.Context, path string) (string, error) {
	key := s.buildPath(path)

	// 如果配置了 URL 前缀，直接使用
	if s.config.URLPrefix != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.config.URLPrefix, "/"), key), nil
	}

	// 生成预签名 URL（15分钟有效期）
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return urlStr, nil
}

// List 列出文件
func (s *S3Storage) List(ctx context.Context, prefix string) ([]*storage.FileInfo, error) {
	keyPrefix := s.buildPath(prefix)

	var files []*storage.FileInfo

	err := s.client.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.Bucket),
		Prefix: aws.String(keyPrefix),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			// 移除前缀，获取相对路径
			relativePath := strings.TrimPrefix(*obj.Key, keyPrefix)
			if relativePath == "" {
				relativePath = *obj.Key
			}

			fileInfo := &storage.FileInfo{
				Name:        filepath.Base(*obj.Key),
				Path:        relativePath,
				Size:        *obj.Size,
				ContentType: storage.MimeType(*obj.Key),
				Extension:   filepath.Ext(*obj.Key),
				URL:         s.buildURL(*obj.Key),
				Hash:        strings.Trim(*obj.ETag, "\""),
				UploadTime:  *obj.LastModified,
				Metadata: map[string]string{
					"bucket":        s.config.Bucket,
					"key":           *obj.Key,
					"etag":          strings.Trim(*obj.ETag, "\""),
					"storage_class": *obj.StorageClass,
				},
			}

			files = append(files, fileInfo)
		}
		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return files, nil
}

// Copy 复制文件
func (s *S3Storage) Copy(ctx context.Context, srcPath, dstPath string) error {
	srcKey := s.buildPath(srcPath)
	dstKey := s.buildPath(dstPath)

	// 构建复制源
	copySource := fmt.Sprintf("%s/%s", s.config.Bucket, srcKey)

	_, err := s.client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.config.Bucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String(copySource),
		ACL:        aws.String(s.getACL()),
	})
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Move 移动文件
func (s *S3Storage) Move(ctx context.Context, srcPath, dstPath string) error {
	// 先复制文件
	err := s.Copy(ctx, srcPath, dstPath)
	if err != nil {
		return fmt.Errorf("failed to copy file during move: %w", err)
	}

	// 删除源文件
	err = s.Delete(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("failed to delete source file during move: %w", err)
	}

	return nil
}

// buildPath 构建对象路径
func (s *S3Storage) buildPath(path string) string {
	// 清理路径
	path = strings.TrimPrefix(path, "/")

	// 如果启用了日期路径
	if s.config.DatePath {
		dateFormat := s.config.DateFormat
		if dateFormat == "" {
			dateFormat = "2006/01/02"
		}
		datePath := time.Now().Format(dateFormat)
		path = datePath + "/" + path
	}

	return path
}

// buildURL 构建文件 URL
func (s *S3Storage) buildURL(key string) string {
	if s.config.URLPrefix != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.config.URLPrefix, "/"), key)
	}

	// 构建默认 URL
	scheme := "https"
	if !s.config.UseSSL {
		scheme = "http"
	}

	if s.config.Endpoint != "" {
		// 自定义端点
		if s.config.PathStyle {
			return fmt.Sprintf("%s://%s/%s/%s", scheme, s.config.Endpoint, s.config.Bucket, key)
		}
		return fmt.Sprintf("%s://%s.%s/%s", scheme, s.config.Bucket, s.config.Endpoint, key)
	}

	// AWS S3 默认端点
	return fmt.Sprintf("%s://%s.s3.%s.amazonaws.com/%s", scheme, s.config.Bucket, s.config.Region, key)
}

// getACL 获取 ACL 设置
func (s *S3Storage) getACL() string {
	if s.config.ACL != "" {
		return s.config.ACL
	}
	return "private"
}
