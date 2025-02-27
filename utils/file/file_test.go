package file

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"testing"
)

func TestGetFileTree(t *testing.T) {
	tree, _ := ReadDirTree("D:\\project\\bit-labs", 2)
	fmt.Println(jsoniter.MarshalToString(tree))
}
