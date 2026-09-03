/**
 * Create Time:2024/1/10
 * User: luchao
 * Email: lcmusic1994@gmail.com
 */

package gcomponent

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qionggemens/gcommon/pkg/glog"
	"github.com/qionggemens/gcommon/pkg/gutil"
)

/*
雪花算法(snowFlake)的具体实现方案:
*/

// snowflakeTwepochMs 固定纪元（毫秒），与 timeGen() 的 UnixMilli 对齐。
// 使用固定值而非 time.Now()，避免进程重启后 ID 空间重置、以及秒/毫秒混用。
const snowflakeTwepochMs int64 = 1577808000000 // 2020-01-01 00:00:00 UTC

type SnowFlake struct {
	mu sync.Mutex
	// 雪花算法开启时的起始时间戳（毫秒）
	twepoch int64

	// 每一部分占用的位数
	workerIdBits     int64 // 每个数据中心的工作机器的编号位数
	datacenterIdBits int64 // 数据中心的编号位数
	sequenceBits     int64 // 每个工作机器每毫秒递增的位数

	// 每一部分最大的数值（含边界，5 bit → 0~31）
	maxWorkerId     int64
	maxDatacenterId int64
	maxSequence     int64

	// 每一部分向左移动的位数
	workerIdShift     int64
	datacenterIdShift int64
	timestampShift    int64

	datacenterId  int64
	workerId      int64
	sequence      int64
	lastTimestamp int64
}

// timeGen 获取当前毫秒时间戳
func (s *SnowFlake) timeGen() int64 {
	return time.Now().UnixMilli()
}

// tilNextMills 自旋等待到 lastTimestamp 的下一毫秒（调用方须已持有 mu）
func (s *SnowFlake) tilNextMills() int64 {
	timeStampMill := s.timeGen()
	for timeStampMill <= s.lastTimestamp {
		time.Sleep(time.Microsecond * 100)
		timeStampMill = s.timeGen()
	}
	return timeStampMill
}

func (s *SnowFlake) NextId() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nowTimestamp := s.timeGen()
	if nowTimestamp < s.lastTimestamp {
		glog.Errorf("NextId fail - msg:%s", fmt.Sprintf("clock moved backwards, Refusing to generate id for %d milliseconds", s.lastTimestamp-nowTimestamp))
		return -1, errors.New("id generate fail")
	}
	if nowTimestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & s.maxSequence
		if s.sequence == 0 {
			nowTimestamp = s.tilNextMills()
		}
	} else {
		s.sequence = 0
	}
	s.lastTimestamp = nowTimestamp
	return (nowTimestamp-s.twepoch)<<s.timestampShift |
			s.datacenterId<<s.datacenterIdShift |
			s.workerId<<s.workerIdShift |
			s.sequence,
		nil
}

func NewSnowFlake(workerId int64, datacenterId int64) (*SnowFlake, error) {
	mySnow := new(SnowFlake)
	mySnow.twepoch = snowflakeTwepochMs
	if workerId < 0 || datacenterId < 0 {
		return nil, errors.New("workerId or datacenterId must not lower than 0 ")
	}

	mySnow.workerIdBits = 5
	mySnow.datacenterIdBits = 5
	mySnow.sequenceBits = 12

	mySnow.maxWorkerId = -1 ^ (-1 << mySnow.workerIdBits)
	mySnow.maxDatacenterId = -1 ^ (-1 << mySnow.datacenterIdBits)
	mySnow.maxSequence = -1 ^ (-1 << mySnow.sequenceBits)

	// 合法范围为 [0, max]，原 >= 会错误拒绝 31
	if workerId > mySnow.maxWorkerId || datacenterId > mySnow.maxDatacenterId {
		return nil, errors.New("workerId or datacenterId must not higher than max value ")
	}
	mySnow.workerIdShift = mySnow.sequenceBits
	mySnow.datacenterIdShift = mySnow.sequenceBits + mySnow.workerIdBits
	mySnow.timestampShift = mySnow.sequenceBits + mySnow.workerIdBits + mySnow.datacenterIdBits

	mySnow.lastTimestamp = -1
	mySnow.workerId = workerId
	mySnow.datacenterId = datacenterId

	return mySnow, nil
}

func NewEasySnowFlake() (*SnowFlake, error) {
	ip4, err := gutil.GetLocalIpx(4)
	if err != nil {
		return nil, err
	}
	strArr := strings.Split(ip4, ".")
	if len(strArr) < 4 {
		return nil, fmt.Errorf("ip4 is invalid [%s]", ip4)
	}
	a, errA := strconv.ParseInt(strArr[0], 10, 64)
	b, errB := strconv.ParseInt(strArr[1], 10, 64)
	c, errC := strconv.ParseInt(strArr[2], 10, 64)
	d, errD := strconv.ParseInt(strArr[3], 10, 64)
	if errA != nil || errB != nil || errC != nil || errD != nil {
		return nil, fmt.Errorf("ip4 parse fail [%s]", ip4)
	}
	m := a*1000 + b
	n := c*1000 + d
	return NewSnowFlake(n%32, m%32)
}
