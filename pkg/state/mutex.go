package state

import (
	"context"
	"log/slog"
	"sync"
	"time"

	slogctx "github.com/veqryn/slog-context"
)

var processingMutex sync.Map // Zero value is empty and ready for use

func LockForProcessing(ctx context.Context) func() {
	slogctx.Debug(ctx, "Waiting for lock")

	key := Provider(ctx) + "/" + ProjectID(ctx) + "/" + MergeRequestID(ctx)

	start := time.Now()
	value, _ := processingMutex.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex) //nolint
	mtx.Lock()

	slogctx.Debug(ctx, "Lock acquired", slog.Duration("waited_for_lock_duration", time.Since(start)))

	return func() { mtx.Unlock() }
}
