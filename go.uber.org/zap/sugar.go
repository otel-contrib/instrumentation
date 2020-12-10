package zap

import "go.uber.org/zap"

// A SugaredLogger wraps the base Logger functionality in a slower, but less
// verbose, API. Any Logger can be converted to a SugaredLogger with its Sugar
// method.
type SugaredLogger = zap.SugaredLogger
