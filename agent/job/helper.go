package job

import (
	"github.com/telemetryapp/gotelemetry"
	"sync"
	"time"
)

// A simple task closure
type PluginHelperClosure func(job *Job)

// A task closure that's associated with a flow
type PluginHelperClosureWithFlow func(job *Job, f *gotelemetry.Flow)

type pluginHelperTask func(job *Job, doneChannel chan bool)

// struct PluginHelper simplifies the process of creating plugins by providing most
// of the required plumbing and allowing the developer to focus on application-specific
// functionality.
//
// When using PluginHelper as the basis for a plugin, you are only required to provide
// an Init() method in which you configure one or more tasks, which can optionally
// be associated with a flow.
//
// PluginHelper will automatically execute tasks asynchronously on a schedule. You
// can, therefore, consider tasks single-purpose and synchronous, performing
// whatever functionality you require and then exiting immediately.
type PluginHelper struct {
	tasks       []pluginHelperTask
	doneChannel chan bool
	waitGroup   *sync.WaitGroup
}

// Creates a new plugin helper and returns it
func NewPluginHelper() *PluginHelper {
	return &PluginHelper{
		tasks:       []pluginHelperTask{},
		doneChannel: make(chan bool, 0),
		waitGroup:   &sync.WaitGroup{},
	}
}

func (e *PluginHelper) addTask(t pluginHelperTask) {
	e.tasks = append(e.tasks, t)
}

// Adds a task to the plugin. The task will be run automarically after the duration specified by
// the interval parameter. Note that interval is measured starting from the end of the last
// execution; therefore, you do not need to worry about conditions like slow networking causing
// successive iterations of a task to “execute over each other.”
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

	e.addTask(t)
}

// Adds a task associated with a flow taken from a map of flows. You can obtain a map of flows by calling
// the MapWidgetToFlows() method of gotelemetry.Board.
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

// Adds a task associated with a flow taken from a board. This method automatically
// handles board prefixes; therefore, you must use the tags exactly as they are defined
// when in the board template.
func (e *PluginHelper) AddTaskWithClosureFromBoardForFlowWithTag(c PluginHelperClosureWithFlow, interval time.Duration, b *gotelemetry.Board, tag string) error {
	flows, err := b.MapWidgetsToFlows()

	if err != nil {
		return err
	}

	return e.AddTaskWithClosureForFlowWithTag(c, interval, flows, tag)
}

// Run method satisfies the requirements of the PluginInstance interface,
// executing all the tasks asynchronously.
func (e *PluginHelper) Run(job *Job) {
	defer e.waitGroup.Done()

	for _, t := range e.tasks {
		e.waitGroup.Add(1)

		go func(t pluginHelperTask) {
			t(job, e.doneChannel)
			e.waitGroup.Done()
		}(t)
	}

	select {
	case <-e.doneChannel:
		return
	}
}

// By default, the plugin helper refuses to reconfigure plugins.
func (e *PluginHelper) Reconfigure(job *Job, config map[string]interface{}) error {
	return gotelemetry.NewError(400, "This pluginc cannot reconfigure itself.")
}

// Terminate waits for all outstanding tasks to be completed and then returns.
func (e *PluginHelper) Terminate(job *Job) {
	e.doneChannel <- true
	e.waitGroup.Wait()
}
