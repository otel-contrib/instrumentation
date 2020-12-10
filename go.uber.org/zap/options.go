package zap

import "go.uber.org/zap"

// An Option configures a Logger.
type Option = zap.Option

// AddCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option). When building wrappers around the
// Logger and SugaredLogger, supplying this Option prevents zap from always
// reporting the wrapper code as the caller.
func AddCallerSkip(skip int) Option {
	return zap.AddCallerSkip(skip)
}
