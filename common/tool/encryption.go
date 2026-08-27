package tool

import (
	"crypto/md5"
	"fmt"
	"io"
)

/** 加密方式 **/

func Md5ByString(str string) string {
	m := md5.New()
	_, _ = io.WriteString(m, str)
	arr := m.Sum(nil)
	return fmt.Sprintf("%x", arr)
}
