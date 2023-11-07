package g

import "time"

// 任务
type GJob interface {
	HandelFunc(*GEngine)
	Expire() time.Duration
}

type gJobServer []GJob

// 开始执行全部任务
func (job gJobServer) start(ge *GEngine) {
	for _, item := range job {
		job.startJob(ge, item)
	}
}
func (job gJobServer) stop() {}

// 开始执行任务
func (job gJobServer) startJob(ge *GEngine, j GJob) {
	time.AfterFunc(j.Expire(), func() {
		j.HandelFunc(ge)
		time.AfterFunc(j.Expire(), func() {
			j.HandelFunc(ge)
		})
	})
}
