package job

import (
	"github.com/telemetryapp/gotelemetry"
	"sync"
	"time"
)

type PluginHelperClosure func(job *Job)
type PluginHelperClosureWithFlow func(job *Job, f *gotelemetry.Flow)

type PluginHelperTask func(job *Job, doneChannel chan bool)

type PluginHelper struct {
	Tasks       []PluginHelperTask
	DoneChannel chan bool
	WaitGroup   *sync.WaitGroup
}

func NewPluginHelper() *PluginHelper {
	return &PluginHelper{
		Tasks:       []PluginHelperTask{},
		DoneChannel: make(chan bool, 0),
		WaitGroup:   &sync.WaitGroup{},
	}
}

func (e *PluginHelper) AddTask(t PluginHelperTask) {
	e.Tasks = append(e.Tasks, t)
}

func (e *PluginHelper) AddTaskWithClosure(c PluginHelperClosure, interval time.Duration) {
	t := func(job *Job, doneChannel chan bool) {
		t := time.NewTimer(interval)
		t.Stop()

		for {
			c(job)

			t.Reset(interval)

			select {
			case <-doneChannel:
				t.Stop()
				return

			case <-t.C:
				break
			}
		}
	}

	e.AddTask(t)
}

func (e *PluginHelper) AddTaskWithClosureForFlowWithTag(c PluginHelperClosureWithFlow, interval time.Duration, flows map[string]*gotelemetry.Flow, tag string) error {
	f, found := flows[tag]

	if !found {
		return gotelemetry.NewError(400, "Flow "+tag+" not found.")
	}

	closure := func(job *Job) {
		c(job, f)
	}

	e.AddTaskWithClosure(closure, interval)

	return nil
}

func (e *PluginHelper) AddTaskWithClosureFromBoardForFlowWithTag(c PluginHelperClosureWithFlow, interval time.Duration, b *gotelemetry.Board, tag string) error {
	flows, err := b.MapWidgetsToFlows()

	if err != nil {
		return err
	}

	return e.AddTaskWithClosureForFlowWithTag(c, interval, flows, tag)
}

func (e *PluginHelper) Run(job *Job) {
	defer e.WaitGroup.Done()

	for _, t := range e.Tasks {
		e.WaitGroup.Add(1)

		go func(t PluginHelperTask) {
			t(job, e.DoneChannel)
			e.WaitGroup.Done()
		}(t)
	}

	select {
	case <-e.DoneChannel:
		return
	}
}

func (e *PluginHelper) Reconfigure(job *Job, config map[string]interface{}) error {
	return gotelemetry.NewError(400, "This pluginc cannot reconfigure itself.")
}

func (e *PluginHelper) Terminate(job *Job) {
	e.DoneChannel <- true
	e.WaitGroup.Wait()
}
