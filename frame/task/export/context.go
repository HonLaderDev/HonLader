package export

import "context"

func (e *ExportTask) startTaskContext() context.Context {
	e.taskMu.Lock()
	defer e.taskMu.Unlock()

	if e.taskCancel != nil {
		e.taskCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.taskCtx = ctx
	e.taskCancel = cancel
	return ctx
}

func (e *ExportTask) finishTaskContext(ctx context.Context) {
	e.taskMu.Lock()
	defer e.taskMu.Unlock()

	if e.taskCtx != ctx {
		return
	}
	e.taskCtx = nil
	e.taskCancel = nil
}

func (e *ExportTask) checkContext() bool {
	e.taskMu.Lock()
	defer e.taskMu.Unlock()

	return e.taskCtx != nil && e.taskCtx.Err() == nil
}

func (e *ExportTask) cancelTask() {
	e.taskMu.Lock()
	defer e.taskMu.Unlock()

	if e.taskCancel != nil {
		e.taskCancel()
	}
}
