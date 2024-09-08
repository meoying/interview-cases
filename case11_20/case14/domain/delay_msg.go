package domain

import "time"

// 延迟消息
type DelayMsg struct {
	ID int64
	// 转发内容
	Value string
	// 转发主题
	Topic string
	// 截止时间
	Deadline time.Time
	Status   DelayMsgStatus
	Ctime    time.Time
	Utime    time.Time
}

type DelayMsgStatus uint8

const (
	// Waiting 待完成
	Waiting DelayMsgStatus = 0
	// Completed 已完成
	Completed DelayMsgStatus = 1
)

func (status DelayMsgStatus) ToUint8() uint8 {
	return uint8(status)
}
