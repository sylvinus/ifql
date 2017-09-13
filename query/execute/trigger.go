package execute

import (
	"fmt"

	"github.com/influxdata/ifql/query"
)

type Trigger interface {
	Triggered(TriggerContext) bool
	Finished() bool
	Reset()
}

type TriggerContext struct {
	Builder               BlockBuilder
	Watermark             Time
	CurrentProcessingTime Time
}

func newTriggerFromSpec(spec query.TriggerSpec) Trigger {
	switch s := spec.(type) {
	case query.AfterWatermarkTriggerSpec:
		return &afterWatermarkTrigger{
			allowedLateness: Duration(s.AllowedLateness),
		}
	case query.RepeatedTriggerSpec:
		return &repeatedlyForever{
			t: newTriggerFromSpec(s.Trigger),
		}
	case query.AfterProcessingTimeTriggerSpec:
		return &afterProcessingTimeTrigger{
			duration: Duration(s.Duration),
		}
	case query.AfterAtLeastCountTriggerSpec:
		return &afterAtLeastCount{
			count: s.Count,
		}
	case query.OrFinallyTriggerSpec:
		return &orFinally{
			main:    newTriggerFromSpec(s.Main),
			finally: newTriggerFromSpec(s.Finally),
		}
	default:
		//TODO(nathanielc): Add proper error handling here.
		// Maybe separate validation of a spec and creation of a spec so we know we cannot error during creation?
		panic(fmt.Sprintf("unsupported trigger spec provided %T", spec))
	}
}

// afterWatermarkTrigger triggers once the watermark is greater than the bounds of the block.
type afterWatermarkTrigger struct {
	allowedLateness Duration
	finished        bool
}

func (t *afterWatermarkTrigger) Triggered(c TriggerContext) bool {
	if c.Watermark >= c.Builder.Bounds().Stop+Time(t.allowedLateness) {
		t.finished = true
	}
	return c.Watermark >= c.Builder.Bounds().Stop
}
func (t *afterWatermarkTrigger) Finished() bool {
	return t.finished
}
func (t *afterWatermarkTrigger) Reset() {
	t.finished = false
}

type repeatedlyForever struct {
	t Trigger
}

func (t *repeatedlyForever) Triggered(c TriggerContext) bool {
	return t.t.Triggered(c)
}
func (t *repeatedlyForever) Finished() bool {
	if t.t.Finished() {
		t.Reset()
	}
	return false
}
func (t *repeatedlyForever) Reset() {
	t.t.Reset()
}

type afterProcessingTimeTrigger struct {
	duration       Duration
	triggerTimeSet bool
	triggerTime    Time
	current        Time
}

func (t *afterProcessingTimeTrigger) Triggered(c TriggerContext) bool {
	if !t.triggerTimeSet {
		t.triggerTimeSet = true
		t.triggerTime = c.CurrentProcessingTime + Time(t.duration)
	}
	t.current = c.CurrentProcessingTime
	return t.current >= t.triggerTime
}
func (t *afterProcessingTimeTrigger) Finished() bool {
	return t.triggerTimeSet && t.current >= t.triggerTime
}
func (t *afterProcessingTimeTrigger) Reset() {
	t.triggerTimeSet = false
}

type afterAtLeastCount struct {
	count int
	rows  int
}

func (t *afterAtLeastCount) Triggered(c TriggerContext) bool {
	t.rows = c.Builder.NRows()
	return t.rows >= t.count
}
func (t *afterAtLeastCount) Finished() bool {
	return t.rows >= t.count
}
func (t *afterAtLeastCount) Reset() {
	t.rows = 0
}

type orFinally struct {
	main     Trigger
	finally  Trigger
	finished bool
}

func (t *orFinally) Triggered(c TriggerContext) bool {
	if t.finally.Triggered(c) {
		t.finished = true
		return true
	}
	return t.main.Triggered(c)
}

func (t *orFinally) Finished() bool {
	return t.finished
}
func (t *orFinally) Reset() {
	t.finished = false
}
