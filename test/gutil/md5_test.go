/**
 * Create Time:2023/4/14
 * User: luchao
 * Email: lcmusic1994@gmail.com
 */

package gutil

import (
	"fmt"
	util "github.com/qionggemens/gcommon/pkg/gutil"
	"testing"
	"time"
)

func TestMd5(t *testing.T) {
	fmt.Println(util.Md5("123"))
	fmt.Println(time.Now().UnixMicro())
	fmt.Println(time.Now().UnixNano())
}
