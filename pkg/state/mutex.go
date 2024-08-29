package state

import (
	"context"
	"sync"
)

var processingMutex sync.Map // Zero value is empty and ready for use

func LockForProcessing(ctx context.Context) func() {
	key := Provider(ctx) + "/" + ProjectID(ctx) + "/" + MergeRequestID(ctx)

	value, _ := processingMutex.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex) //nolint
	mtx.Lock()

	return func() { mtx.Unlock() }
}
