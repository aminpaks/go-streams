package xredis

type XSortedQueueOptions struct {
	MaxRetries int
	Consuming  int64
	Consumers  int
}

func NewXSortedQueueOptions() *XSortedQueueOptions {
	return &XSortedQueueOptions{
		MaxRetries: 3,
		Consuming:  10,
		Consumers:  1,
	}
}

func (x *XSortedQueueOptions) Initialize() {
	if x.Consuming < 1 {
		x.Consuming = 1
	}
	if x.Consumers < 1 {
		x.Consumers = 1
	}
}
