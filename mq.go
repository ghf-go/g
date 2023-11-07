package g

import "time"

// 消息队列
type GMQ interface {
	Before(*GEngine)
	Handler()
	After()
}
type gmqServer []GMQ

// 开始执行全部任务
func (mq gmqServer) start(ge *GEngine) {
	ch := make(chan GMQ, 1)
	for _, qq := range mq {
		go mq.startMq(ge, qq, ch)
	}
	for {
		q := <-ch
		go mq.startMq(ge, q, ch)
	}
}
func (mq gmqServer) stop() {}

// 启动队列
func (mq gmqServer) startMq(ge *GEngine, q GMQ, ch chan GMQ) {
	if e := recover(); e != nil {
		ch <- q
	}
	defer q.After()
	q.Before(ge)
	q.Handler()
}

// redis 队列实现
type GmqRedis struct {
	ge       *GEngine
	MsgFunc  func(msg string)
	RedisKey string
}

// 新建Redis的消息对垒
func NewMqRedis(redisKey string, msgcall func(msg string)) *GmqRedis {
	return &GmqRedis{
		MsgFunc:  msgcall,
		RedisKey: redisKey,
	}
}
func (r *GmqRedis) Before(ge *GEngine) {
	r.ge = ge
}
func (r *GmqRedis) Handler() {
	for {
		msgs, _ := r.ge.GetRedis().BLPop(r.ge.Ctx, time.Second*30, r.RedisKey).Result()
		for _, msg := range msgs {
			r.MsgFunc(msg)
		}
	}
}
func (r *GmqRedis) After() {}
