package file

import (
	"fmt"
	"time"
)

type FileTransfer struct {
	Path     string  `json:"path"`     // 文件名
	Size     int64   `json:"size"`     // 文件大小
	Progress float64 `json:"progress"` // 进度
	Status   string  `json:"status"`
	Error    string  `json:"error"`
}

type Node struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ParentPath string    `json:"parentPath"`
	IsDir      bool      `json:"isDir"`
	Children   []*Node   `json:"children,omitempty"`
	Size       int64     `json:"size"`
	SizeInfo   string    `json:"sizeInfo"`
	ModTime    time.Time `json:"modTime"`
}

func (ft *Node) HumanSize() string {
	b := ft.Size
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
