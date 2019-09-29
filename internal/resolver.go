package internal

import (
	"github.com/icelolly/go-errors"
	"github.com/seeruk/bedrockd/internal/bedrock"
	"go.uber.org/zap"
)

// Resolver is a plain Go type that's used for wiring application dependencies together. It's
// exposed API allows callers to "resolve" application dependencies. Resolving a dependency may call
// other resolver methods to resolve dependencies.
type Resolver struct {
	process *bedrock.Process
}

// NewResolver returns a new Resolver instance.
func NewResolver() *Resolver {
	return &Resolver{}
}

// ResolveBedrockProcess will return a singleton instance of the bedrock process.
func (r *Resolver) ResolveBedrockProcess() *bedrock.Process {
	if r.process == nil {
		r.process = bedrock.NewProcess(r.ResolveLogger())
	}

	return r.process
}

// ResolveLogger returns this application's logger instance.
func (r *Resolver) ResolveLogger() *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}

	logger, err := config.Build()
	if err != nil {
		errors.Fatal(err)
	}

	return logger.Sugar()
}
