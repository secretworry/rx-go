package rx

import "context"

func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func toError(e interface{}) error {
	if err, ok := e.(error); ok {
		return err
	}
	return ErrPanic(e)
}
