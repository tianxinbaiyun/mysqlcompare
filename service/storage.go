package service

import (
	"log"
	"sync/atomic"
)

// 分配存储管道
var (
	DataChan  = make(chan *Data, 5000) // 数据队列
	DataCount = int64(0)
)

// DataConsume 数据消费
func DataConsume() {
	for {
		select {
		// 从存储队列取出数据，进行存储
		case s := <-DataChan:
			go s.Operation()
		}
	}
}

// Data 数据结构体
type Data struct {
	Unique   string   // 唯一值
	Row      []string // 数据库记录
	fields   []string // 不一致的字段
	TryTimes int      // 处理次数
}

// Put 把数据提交到，数据队列里
func (s *Data) Put() {
	DataChan <- s
}

// Operation 保存数据数据
func (s *Data) Operation() {
	atomic.AddInt64(&DataCount, 1)
	log.Println(DataCount, s.Unique, s.fields)

}
