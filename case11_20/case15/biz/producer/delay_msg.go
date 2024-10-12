package producer

// DelayMsg 延迟消息
type DelayMsg struct {
	// 转发内容
	Value []byte
	// 代表这个消息的唯一 ID，我们会用来在延迟消费者这边去重
	// 如果你不传递这个字段，我们就不会执行去重
	// 但是不管去不去重，你的消费者都必须是幂等
	// 因为延迟发送者这边，无法做到只发送一次
	Key string
	// 转发主题，也就是业务主题
	Topic string
	// 到什么时候发出去
	Deadline int64
}
