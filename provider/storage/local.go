package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	config *LocalConfig
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(config *LocalConfig) *LocalStorage {
	return &LocalStorage{
		config: config,
	}
}

// Put 上传文件
func (ls *LocalStorage) Put(ctx context.Context, path string, reader io.Reader, size int64) (*FileInfo, error) {
	// 构建完整路径
	fullPath := ls.buildPath(path)

	// 创建目录
	if ls.config.CreateDirs {
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, os.FileMode(ls.config.DirMode)); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 创建文件
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(ls.config.FileMode))
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 计算文件哈希
	hash := md5.New()
	multiWriter := io.MultiWriter(file, hash)

	// 复制数据
	written, err := io.Copy(multiWriter, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// 获取文件信息
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %w", err)
	}

	// 构建文件信息
	fileInfo := &FileInfo{
		Name:       filepath.Base(path),
		Path:       path,
		Size:       written,
		Extension:  filepath.Ext(path),
		URL:        ls.buildURL(path),
		Hash:       fmt.Sprintf("%x", hash.Sum(nil)),
		UploadTime: stat.ModTime(),
		Metadata:   make(map[string]string),
	}

	return fileInfo, nil
}

// PutFile 上传本地文件
func (ls *LocalStorage) PutFile(ctx context.Context, path string, localPath string) (*FileInfo, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %w", err)
	}

	return ls.Put(ctx, path, file, stat.Size())
}

// Get 获取文件
func (ls *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := ls.buildPath(path)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete 删除文件
func (ls *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := ls.buildPath(path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (ls *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := ls.buildPath(path)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Size 获取文件大小
func (ls *LocalStorage) Size(ctx context.Context, path string) (int64, error) {
	fullPath := ls.buildPath(path)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("file not found: %s", path)
		}
		return 0, fmt.Errorf("failed to get file stat: %w", err)
	}

	return stat.Size(), nil
}

// URL 获取文件访问 URL
func (ls *LocalStorage) URL(ctx context.Context, path string) (string, error) {
	return ls.buildURL(path), nil
}

// List 列出文件
func (ls *LocalStorage) List(ctx context.Context, prefix string) ([]*FileInfo, error) {
	var files []*FileInfo

	prefixPath := ls.buildPath(prefix)

	err := filepath.Walk(prefixPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(ls.config.Root, path)
		if err != nil {
			return err
		}

		// 标准化路径分隔符
		relPath = filepath.ToSlash(relPath)

		fileInfo := &FileInfo{
			Name:       info.Name(),
			Path:       relPath,
			Size:       info.Size(),
			Extension:  filepath.Ext(info.Name()),
			URL:        ls.buildURL(relPath),
			UploadTime: info.ModTime(),
			Metadata:   make(map[string]string),
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// Copy 复制文件
func (ls *LocalStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	srcFullPath := ls.buildPath(srcPath)
	dstFullPath := ls.buildPath(dstPath)

	// 打开源文件
	srcFile, err := os.Open(srcFullPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// 创建目标目录
	if ls.config.CreateDirs {
		dir := filepath.Dir(dstFullPath)
		if err := os.MkdirAll(dir, os.FileMode(ls.config.DirMode)); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 创建目标文件
	dstFile, err := os.OpenFile(dstFullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(ls.config.FileMode))
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// 复制数据
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Move 移动文件
func (ls *LocalStorage) Move(ctx context.Context, srcPath, dstPath string) error {
	srcFullPath := ls.buildPath(srcPath)
	dstFullPath := ls.buildPath(dstPath)

	// 创建目标目录
	if ls.config.CreateDirs {
		dir := filepath.Dir(dstFullPath)
		if err := os.MkdirAll(dir, os.FileMode(ls.config.DirMode)); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 移动文件
	if err := os.Rename(srcFullPath, dstFullPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// buildPath 构建完整路径
func (ls *LocalStorage) buildPath(path string) string {
	// 清理路径
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")

	dateFormat := strings.TrimSpace(ls.config.DateFormat)
	if dateFormat != "" {
		datePath := time.Now().Format(normalizeDateFormat(dateFormat))
		path = filepath.Join(datePath, path)
	}

	return filepath.Join(ls.config.Root, path)
}

// buildURL 构建访问 URL
func (ls *LocalStorage) buildURL(path string) string {
	// 清理路径
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")
	path = filepath.ToSlash(path)

	dateFormat := strings.TrimSpace(ls.config.DateFormat)
	if dateFormat != "" {
		datePath := time.Now().Format(normalizeDateFormat(dateFormat))
		path = datePath + "/" + path
	}

	urlPrefix := strings.TrimSuffix(ls.config.URLPrefix, "/")
	return urlPrefix + "/" + path
}
