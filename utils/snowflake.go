package utils

import (
	"errors"
	"sync"
	"time"
)

/*
 * 算法解释
 * SnowFlake的结构如下(每部分用-分开):<br>
 * 0 - 0000000000 0000000000 0000000000 0000000000 0 - 00000 - 00000 - 000000000000 <br>
 * 1位标识，由于long基本类型在Java中是带符号的，最高位是符号位，正数是0，负数是1，所以id一般是正数，最高位是0<br>
 * 41位时间截(毫秒级)，注意，41位时间截不是存储当前时间的时间截，而是存储时间截的差值（当前时间截 - 开始时间截)
 * 得到的值），这里的的开始时间截，一般是我们的id生成器开始使用的时间，由我们程序来指定的（如下的epoch属性）。
 * 41位的时间截，可以使用69年，年T = (1L << 41) / (1000L * 60 * 60 * 24 * 365) = 69<br>
 * 10位的数据机器位，可以部署在1024个节点，包括5位datacenterId和5位workerId<br>
 * 12位序列，毫秒内的计数，12位的计数顺序号支持每个节点每毫秒(同一机器，同一时间截)产生4096个ID序号<br>
 * 加起来刚好64位，为一个Long型。<br>
 * SnowFlake的优点是，整体上按照时间自增排序，并且整个分布式系统内不会产生ID碰撞(由数据中心ID和机器ID作区分)，并且效率较高，经测试，SnowFlake每秒能够产生26万ID左右。
 */

const (
	workerIDBits     = uint64(5) // 10bit 工作机器ID中的 5bit workerID
	dataCenterIDBits = uint64(5) // 10 bit 工作机器ID中的 5bit dataCenterID
	sequenceBits     = uint64(12)

	maxWorkerID     = int64(-1) ^ (int64(-1) << workerIDBits) //节点ID的最大值 用于防止溢出
	maxDataCenterID = int64(-1) ^ (int64(-1) << dataCenterIDBits)
	maxSequence     = int64(-1) ^ (int64(-1) << sequenceBits)

	timeLeft = uint8(22) // timeLeft = workerIDBits + sequenceBits // 时间戳向左偏移量
	dataLeft = uint8(17) // dataLeft = dataCenterIDBits + sequenceBits
	workLeft = uint8(12) // workLeft = sequenceBits // 节点IDx向左偏移量
	// 2020-10-10 00:00:00 +0800 CST
	twepoch = int64(1602259200000) // 常量时间戳(毫秒)
)

type idGenerator struct {
	// 雪花算法需要
	w *Worker
}

var i *idGenerator
var once sync.Once
var initOnce sync.Once

func GetInstance() *idGenerator {
	once.Do(func() {
		i = &idGenerator{}
	})
	return i
}

func InitSnowFlakeWorker(workerID, dataCenterID int64) error {
	initOnce.Do(func() {
		GetInstance().w = NewWorker(workerID, dataCenterID)
	})

	w := GetInstance().w
	if w == nil {
		return errors.New("snowflake worker not initialized")
	}
	if w.WorkerID != workerID || w.DataCenterID != dataCenterID {
		return errors.New("snowflake worker already initialized with different ids")
	}
	return nil
}

type Worker struct {
	mu           sync.Mutex // 线程互斥锁
	LastStamp    int64      // 记录上一次ID的时间戳
	WorkerID     int64      // 该节点的ID
	DataCenterID int64      // 该节点的 数据中心ID
	Sequence     int64      // 当前毫秒已经生成的ID序列号(从0 开始累加) 1毫秒内最多生成4096个ID
}

// 分布式情况下,我们应通过外部配置文件或其他方式为每台机器分配独立的id
func NewWorker(workerID, dataCenterID int64) *Worker {
	if workerID > maxWorkerID {
		workerID = maxWorkerID
	}

	if dataCenterID > maxDataCenterID {
		dataCenterID = maxDataCenterID
	}

	return &Worker{
		WorkerID:     workerID,
		LastStamp:    0,
		Sequence:     0,
		DataCenterID: dataCenterID,
	}
}

/*
*
获取当前时间的时间戳
@return int64 当前时间戳
*/
func (w *Worker) getMilliSeconds() int64 {
	return time.Now().UnixMilli()
}

/*
*
通过雪花算法获取id
@return int64 生成的id
@return error 返回的错误
*/
func SnowFlakeNextID() (int64, error) {
	w := GetInstance().w
	if w == nil {
		return 0, errors.New("snowflake worker not initialized")
	}
	return w.NextID()
}

/*
*
具体的雪花算法实现
@return int64 生成的id
@return error 返回的错误
*/
func (w *Worker) NextID() (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	timeStamp := w.getMilliSeconds()
	if timeStamp < w.LastStamp {
		return 0, errors.New("time is moving backwards,waiting until")
	}

	if w.LastStamp == timeStamp {

		w.Sequence = (w.Sequence + 1) & maxSequence

		if w.Sequence == 0 {
			for timeStamp <= w.LastStamp {
				timeStamp = w.getMilliSeconds()
			}
		}
	} else {
		w.Sequence = 0
	}

	w.LastStamp = timeStamp
	id := ((timeStamp - twepoch) << timeLeft) |
		(w.DataCenterID << dataLeft) |
		(w.WorkerID << workLeft) |
		w.Sequence

	return id, nil
}
