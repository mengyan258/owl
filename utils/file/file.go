package file

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// NormalizedPath 归一化处理路径，去除路径中错误的 \\ //
func NormalizedPath(path string) string {
	pattern := `[\\/]{1,}`
	re := regexp.MustCompile(pattern)        // 编译正则表达式
	result := re.ReplaceAllString(path, "/") // 使用正则表达式替换

	result = strings.TrimRight(result, "/") // 去除末尾 /
	return result
}

func CreateDirIfNotExists(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0777)
	}
}

func DirIsEmpty(dirPath string) bool {

	// 获取文件夹中的文件列表
	fileList, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	// 检查文件列表是否为空
	if len(fileList) == 0 {
		return true
	} else {
		return false
	}
}

// ReadDirTree 读取目录结构
// rootPath: 根路径
// maxLevel: 最大遍历层级（0表示不限制）
func ReadDirTree(rootPath string, maxLevel int) (*Node, error) {
	// 统一路径分隔符
	rootPath = filepath.Clean(rootPath)

	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}

	parentPath := filepath.Dir(rootPath)

	isSymlink := info.Mode()&os.ModeSymlink != 0

	node := &Node{
		Name:       info.Name(),
		Path:       rootPath,
		IsDir:      info.IsDir(),
		Size:       info.Size(),    // 文件大小
		ModTime:    info.ModTime(), // 修改时间
		IsSymlink:  isSymlink,
		ParentPath: normalizePathSeparator(parentPath),
	}
	node.SizeInfo = node.HumanSize()

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return node, err
	}
	// 添加排序逻辑
	sort.Slice(entries, func(i, j int) bool {
		// 目录优先
		if entries[i].IsDir() && !entries[j].IsDir() {
			return true
		}
		if !entries[i].IsDir() && entries[j].IsDir() {
			return false
		}
		// 相同类型时按文件名排序
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		childPath := filepath.Join(rootPath, entry.Name())
		fileInfo, err := entry.Info()
		if err != nil {
			return node, err
		}

		isSymlink = fileInfo.Mode()&os.ModeSymlink != 0
		childNode := &Node{
			Name:       entry.Name(),
			Path:       childPath,
			ParentPath: rootPath,
			IsDir:      entry.IsDir(),
			Size:       fileInfo.Size(),
			SizeInfo:   "",
			ModTime:    fileInfo.ModTime(),
			IsSymlink:  isSymlink,
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

// 统一路径分隔符为当前系统的分隔符
func normalizePathSeparator(path string) string {
	return strings.ReplaceAll(path, "/", string(filepath.Separator))
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ReadFileSafely 安全文本文件读取方法
func ReadFileSafely(path string, maxSize int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 检查文件大小
	fi, err := file.Stat()
	if err != nil {
		return "", err
	}
	if fi.Size() > maxSize {
		return "", fmt.Errorf("file exceeds max size: %d bytes", maxSize)
	}

	// 读取前512字节用于检测文件类型
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// 检测文件类型
	mimeType := http.DetectContentType(buffer[:n])
	if !isAllowedTextType(mimeType) {
		return "", errors.New("只能打开文本文件")
	}

	// 重置文件指针到开头
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// 使用缓冲读取
	buf := new(strings.Builder)
	_, err = io.CopyN(buf, file, maxSize)
	if err != nil && err != io.EOF {
		return "", err
	}

	return buf.String(), nil
}
