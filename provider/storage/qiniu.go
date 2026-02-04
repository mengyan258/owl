package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// QiniuStorage 七牛云存储实现
type QiniuStorage struct {
	mac           *qbox.Mac
	config        *QiniuConfig
	bucketManager *storage.BucketManager
	uploader      *storage.FormUploader
	cfg           *storage.Config
}

// NewQiniuStorage 创建七牛云存储实例
func NewQiniuStorage(config *QiniuConfig) (*QiniuStorage, error) {
	// 创建认证对象
	mac := qbox.NewMac(config.AccessKey, config.SecretKey)

	// 创建配置对象
	cfg := &storage.Config{
		UseHTTPS:      config.UseSSL,
		UseCdnDomains: false,
	}

	// 设置区域
	if config.Zone != "" {
		switch config.Zone {
		case "z0", "华东":
			cfg.Zone = &storage.ZoneHuadong
		case "z1", "华北":
			cfg.Zone = &storage.ZoneHuabei
		case "z2", "华南":
			cfg.Zone = &storage.ZoneHuanan
		case "na0", "北美":
			cfg.Zone = &storage.ZoneBeimei
		case "as0", "新加坡":
			cfg.Zone = &storage.ZoneXinjiapo
		case "华东浙江2区":
			cfg.Zone = &storage.ZoneHuadongZheJiang2
		default:
			cfg.Zone = &storage.ZoneHuadong // 默认华东
		}
	} else {
		cfg.Zone = &storage.ZoneHuadong // 默认华东
	}

	// 创建存储桶管理器
	bucketManager := storage.NewBucketManager(mac, cfg)

	// 创建上传器
	uploader := storage.NewFormUploader(cfg)

	return &QiniuStorage{
		mac:           mac,
		config:        config,
		bucketManager: bucketManager,
		uploader:      uploader,
		cfg:           cfg,
	}, nil
}

// Put 上传文件
func (q *QiniuStorage) Put(ctx context.Context, path string, reader io.Reader, size int64) (*FileInfo, error) {
	key := q.buildPath(path)

	// 读取数据并计算 MD5
	buf := new(bytes.Buffer)
	hash := md5.New()
	tee := io.TeeReader(reader, hash)

	_, err := buf.ReadFrom(tee)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// 生成上传凭证
	putPolicy := storage.PutPolicy{
		Scope: q.config.Bucket,
	}
	upToken := putPolicy.UploadToken(q.mac)

	// 上传文件
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}

	err = q.uploader.Put(ctx, &ret, upToken, key, bytes.NewReader(buf.Bytes()), int64(buf.Len()), &putExtra)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 构建文件信息
	fileInfo := &FileInfo{
		Name:        filepath.Base(path),
		Path:        path,
		Size:        int64(buf.Len()),
		ContentType: MimeType(path),
		Extension:   filepath.Ext(path),
		URL:         q.buildURL(key),
		Hash:        fmt.Sprintf("%x", hash.Sum(nil)),
		UploadTime:  time.Now(),
		Metadata: map[string]string{
			"bucket": q.config.Bucket,
			"key":    key,
			"hash":   ret.Hash,
		},
	}

	return fileInfo, nil
}

// PutFile 上传本地文件
func (q *QiniuStorage) PutFile(ctx context.Context, path string, localPath string) (*FileInfo, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %w", err)
	}

	return q.Put(ctx, path, file, stat.Size())
}

// Get 获取文件
func (q *QiniuStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	key := q.buildPath(path)
	url := q.buildURL(key)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to get file, status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// Delete 删除文件
func (q *QiniuStorage) Delete(ctx context.Context, path string) error {
	key := q.buildPath(path)

	err := q.bucketManager.Delete(q.config.Bucket, key)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (q *QiniuStorage) Exists(ctx context.Context, path string) (bool, error) {
	key := q.buildPath(path)

	_, err := q.bucketManager.Stat(q.config.Bucket, key)
	if err != nil {
		// 检查是否是文件不存在的错误
		if strings.Contains(err.Error(), "no such file or directory") ||
			strings.Contains(err.Error(), "612") { // 七牛云文件不存在错误码
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// Size 获取文件大小
func (q *QiniuStorage) Size(ctx context.Context, path string) (int64, error) {
	key := q.buildPath(path)

	fileInfo, err := q.bucketManager.Stat(q.config.Bucket, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return fileInfo.Fsize, nil
}

// URL 获取文件访问 URL
func (q *QiniuStorage) URL(ctx context.Context, path string) (string, error) {
	key := q.buildPath(path)
	return q.buildURL(key), nil
}

// List 列出文件
func (q *QiniuStorage) List(ctx context.Context, prefix string) ([]*FileInfo, error) {
	keyPrefix := q.buildPath(prefix)

	var files []*FileInfo
	marker := ""
	limit := 1000

	for {
		entries, _, nextMarker, hasNext, err := q.bucketManager.ListFiles(q.config.Bucket, keyPrefix, "", marker, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, entry := range entries {
			// 移除前缀，获取相对路径
			relativePath := strings.TrimPrefix(entry.Key, keyPrefix)
			if relativePath == "" {
				relativePath = entry.Key
			}

			fileInfo := &FileInfo{
				Name:        filepath.Base(entry.Key),
				Path:        relativePath,
				Size:        entry.Fsize,
				ContentType: entry.MimeType,
				Extension:   filepath.Ext(entry.Key),
				URL:         q.buildURL(entry.Key),
				Hash:        entry.Hash,
				UploadTime:  time.Unix(entry.PutTime/10000000, 0), // 七牛云时间戳是纳秒级别
				Metadata: map[string]string{
					"bucket":    q.config.Bucket,
					"key":       entry.Key,
					"hash":      entry.Hash,
					"mime_type": entry.MimeType,
					"end_user":  entry.EndUser,
				},
			}

			files = append(files, fileInfo)
		}

		if !hasNext {
			break
		}
		marker = nextMarker
	}

	return files, nil
}

// Copy 复制文件
func (q *QiniuStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	srcKey := q.buildPath(srcPath)
	dstKey := q.buildPath(dstPath)

	err := q.bucketManager.Copy(q.config.Bucket, srcKey, q.config.Bucket, dstKey, true)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Move 移动文件
func (q *QiniuStorage) Move(ctx context.Context, srcPath, dstPath string) error {
	srcKey := q.buildPath(srcPath)
	dstKey := q.buildPath(dstPath)

	err := q.bucketManager.Move(q.config.Bucket, srcKey, q.config.Bucket, dstKey, true)
	if err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// buildPath 构建对象路径
func (q *QiniuStorage) buildPath(path string) string {
	// 清理路径
	path = strings.TrimPrefix(path, "/")

	// 如果启用了日期路径
	if q.config.DatePath {
		dateFormat := q.config.DateFormat
		if dateFormat == "" {
			dateFormat = "2006/01/02"
		}
		datePath := time.Now().Format(dateFormat)
		path = datePath + "/" + path
	}

	return path
}

// buildURL 构建文件 URL
func (q *QiniuStorage) buildURL(key string) string {
	scheme := "http"
	if q.config.UseSSL {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s", scheme, q.config.Domain, key)
}
