package postal

import (
	"sync"
	"time"

	"github.com/tomygin/box/log"
)

type Msger interface {
	Init() bool
	Send(title, msg string) bool
}

// psotal保存所有告警消息的客户端
type postal struct {
	msgers []Msger
	done   chan bool
}

// NewPostal会根据当前给的配置信息去初始化每个告警客户端
// 如果满足Msger接口并且初始化成功才会添加到postal
func NewPostal(configMsgers ...interface{}) *postal {
	postal := &postal{msgers: make([]Msger, 0, 3), done: make(chan bool, 1)}
	for _, msger := range configMsgers {
		if m, ok := msger.(Msger); ok {
			if m.Init() {
				postal.msgers = append(postal.msgers, m)
			}
		}
	}
	return postal
}

// Send控制所有告警客户端发送告警信息
func (p *postal) Send(title, msg string) {

	var allTask sync.WaitGroup

	taskNums := len(p.msgers)
	allTask.Add(taskNums)

	//检查是否完成
	go func(done chan<- bool, w *sync.WaitGroup) {

		//开始等待
		w.Wait()
		p.done <- true
		// close(p.done)

	}(p.done, &allTask)

	for _, msger := range p.msgers {

		go func(m Msger, w *sync.WaitGroup) {
			m.Send(title, msg)
			w.Done()

		}(msger, &allTask)

	}

	// 等待完成
	// 平均最大给每个任务1秒的时间
	timeout := time.Duration(taskNums*1) * time.Second
	select {
	case <-p.done:
		break
	case <-time.After(timeout):
		//发送超时要更新 done chan bool
		//原来的就交给GC回收
		p.done = make(chan bool, 1)
		log.Error("msger send msg timeout")
	}

}

// Shutdown会等待所有的告警信息发送完成后再退出
// timeout是最长等待告警信息发送完成的时间
// func (p *postal) Shutdown(timeout time.Duration) {

// 	//延时等待完成
// 	select {
// 	case <-p.done:
// 		break
// 	case <-time.After(timeout):
// 		if timeout == 0 {
// 			log.Info("msger force quit suceess")
// 		} else {
// 			log.Error("msger send msg Timeout")
// 		}
// 	}

// 	for _, msger := range p.msgers {
// 		msger.Shutdown()
// 	}
// }
