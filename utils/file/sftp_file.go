package file

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// NormalizedRemotePath 归一化远程路径处理
func NormalizedRemotePath(remotePath string) string {
	pattern := `[\\/]{1,}`
	re := regexp.MustCompile(pattern)
	result := re.ReplaceAllString(remotePath, "/")
	return strings.TrimRight(result, "/")
}

// GetFolderChild 读取远程目录结构
func GetFolderChild(client *sftp.Client, fullPath string) (*Node, error) {
	info, err := client.Stat(fullPath)
	if err != nil {
		return nil, nil
	}

	parentPath := filepath.Dir(fullPath)

	isSymlink := info.Mode()&os.ModeSymlink != 0
	node := &Node{
		Name:       info.Name(),
		Path:       fullPath,
		IsDir:      info.IsDir(),
		Size:       info.Size(),
		ModTime:    info.ModTime(),
		IsSymlink:  isSymlink,
		ParentPath: NormalizedRemotePath(parentPath),
	}
	node.SizeInfo = node.HumanSize()

	if node.IsDir {
		entries, err := client.ReadDir(fullPath)
		if err != nil {
			return node, err
		}

		// 排序：目录在前，文件在后；同类型按名称排序
		sort.Slice(entries, func(i, j int) bool {
			a, b := entries[i], entries[j]
			if a.IsDir() && !b.IsDir() {
				return true
			}
			if !a.IsDir() && b.IsDir() {
				return false
			}
			return a.Name() < b.Name()
		})

		for _, entry := range entries {
			childPath := filepath.ToSlash(filepath.Join(fullPath, entry.Name()))
			isSymlink = entry.Mode()&os.ModeSymlink != 0
			childNode := &Node{
				Name:       entry.Name(),
				Path:       childPath,
				IsDir:      entry.IsDir() || isSymlink,
				Size:       entry.Size(),
				ModTime:    entry.ModTime(),
				ParentPath: NormalizedRemotePath(fullPath),
				IsSymlink:  isSymlink,
			}
			childNode.SizeInfo = childNode.HumanSize()

			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}

// ReadRemoteFileSafely 安全读取远程文本文件
func ReadRemoteFileSafely(client *sftp.Client, filePath string, maxSize int64) (string, error) {
	// 打开远程文件
	file, err := client.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开远程文件失败: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取远程文件信息失败: %w", err)
	}

	// 检查文件大小
	if info.Size() > maxSize {
		return "", fmt.Errorf("文件太大无法打开: %d bytes", maxSize)
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("读取文件头信息失败: %w", err)
	}

	// 检测文件类型
	mimeType := http.DetectContentType(buffer[:n])
	if !isAllowedTextType(mimeType) {
		return "", fmt.Errorf("此文件类型不允许打开: %s", mimeType)
	}

	// 重置文件指针到开头
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// 使用缓冲读取
	buf := new(strings.Builder)
	if _, err := io.CopyN(buf, file, maxSize); err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return buf.String(), nil
}

// isAllowedTextType 检查是否是允许的文本类型
func isAllowedTextType(mimeType string) bool {
	allowedTypes := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/yaml",
		"application/x-yaml",
		"application/x-www-form-urlencoded",
	}

	for _, t := range allowedTypes {
		if strings.HasPrefix(mimeType, t) {
			return true
		}
	}
	return false
}
func SaveFile(client *sftp.Client, filePath, content string) error {
	// 创建一个缓冲区，将字符串内容放入缓冲区
	buffer := bytes.NewBufferString(content)

	// 打开或创建文件，如果文件存在则覆盖
	file, err := client.Create(filePath)
	if err != nil {

		return err
	}
	defer file.Close()

	// 将缓冲区的内容写入文件
	_, err = io.Copy(file, buffer)
	if err != nil {

		return err
	}

	return nil
}

// UploadFile 上传文件
func UploadFile(client *sftp.Client, localFilePath, remoteFilePath string) {
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	dstFile, err := client.Create(remoteFilePath)
	if err != nil {
		return
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return
}

type TransferInfo struct {
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
}

// 递归获取所有文件
func GetFiles(client *sftp.Client, localPath, remotePath string) ([]TransferInfo, error) {
	dirName := filepath.Base(localPath) // 上传的文件夹名称

	var files []TransferInfo
	walkErr := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		targetDir := filepath.Join(remotePath, dirName)
		targetPath := filepath.ToSlash(filepath.Join(targetDir, relPath))

		if info.IsDir() {
			// 创建远程目录
			if err := client.MkdirAll(targetPath); err != nil {
				return err
			}
		} else {
			files = append(files, TransferInfo{
				LocalPath:  path,
				RemotePath: targetPath,
			})
		}

		return nil
	})

	return files, walkErr
}

// RecursiveGetFolderFiles 递归获取文件夹中的所有文件
func RecursiveGetFolderFiles(client *sftp.Client, remotePath, localPath string) ([]TransferInfo, error) {
	// 读取远程目录结构
	remoteTree, err := GetFolderChild(client, remotePath)
	if err != nil {
		return nil, err
	}

	return recursiveGetFolderFiles(client, remoteTree, localPath)
}

func recursiveGetFolderFiles(client *sftp.Client, remoteTree *Node, localPath string) ([]TransferInfo, error) {
	// 构建本地目录路径
	localDirPath := filepath.Join(localPath, remoteTree.Name)

	var trans []TransferInfo
	if remoteTree.IsDir {
		// 创建本地目录
		if err := os.MkdirAll(localDirPath, os.ModePerm); err != nil {
			return nil, err
		}

		// 递归下载子目录和文件
		for _, child := range remoteTree.Children {
			if child.IsDir {
				tx, err := RecursiveGetFolderFiles(client, child.Path, localDirPath)
				if err != nil {
					return nil, err
				}
				trans = append(trans, tx...)
			} else {
				trans = append(trans, TransferInfo{
					LocalPath:  filepath.Join(localDirPath, child.Name),
					RemotePath: child.Path,
				})
			}
		}

	}
	return trans, nil
}

// DownloadFile 下载远程文件到本地
func DownloadFile(client *sftp.Client, remoteFilePath, localFilePath string) error {

	// 打开远程文件
	remoteFile, err := client.Open(remoteFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFilePath = filepath.Join(localFilePath, filepath.Base(remoteFilePath))
	// 创建本地文件
	localFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// 将远程文件内容复制到本地文件
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return err
	}

	return nil
}

// RemoteRemoveAll 递归删除远程目录及其中的所有内容
func RemoteRemoveAll(sftpClient *sftp.Client, path string) error {
	// 检查路径是否存在
	_, err := sftpClient.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("路径不存在: %s", path)
		}
		return fmt.Errorf("无法访问路径 %s: %v", path, err)
	}

	// 遍历目录并递归删除子项
	err = walkAndDelete(sftpClient, path, path)
	if err != nil {
		return fmt.Errorf("删除路径内容时出错 %s: %v", path, err)
	}

	// 最后删除空目录
	err = sftpClient.RemoveDirectory(path)
	if err != nil {
		return fmt.Errorf("无法删除目录 %s: %v", path, err)
	}

	return nil
}

// walkAndDelete 递归遍历并删除目录中的所有内容
func walkAndDelete(sftpClient *sftp.Client, rootPath, currentPath string) error {
	files, err := sftpClient.ReadDir(currentPath)
	if err != nil {
		return fmt.Errorf("无法读取目录 %s: %v", currentPath, err)
	}

	for _, file := range files {
		filePath := sftpClient.Join(currentPath, file.Name())

		if file.IsDir() {
			// 如果是目录，递归删除
			err = walkAndDelete(sftpClient, rootPath, filePath)
			if err != nil {
				return err
			}
			// 删除空目录
			err = sftpClient.RemoveDirectory(filePath)
			if err != nil {
				return fmt.Errorf("无法删除目录 %s: %v", filePath, err)
			}
		} else {
			// 如果是文件，直接删除
			err = sftpClient.Remove(filePath)
			if err != nil {
				return fmt.Errorf("无法删除文件 %s: %v", filePath, err)
			}
		}
	}

	return nil
}
