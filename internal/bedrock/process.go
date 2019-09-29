package bedrock

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"

	"github.com/icelolly/go-errors"
	"go.uber.org/zap"
)

// Process provides a wrapper around the bedrock server process. It handles interaction with input
// and output, and also the lifecycle of the bedrock server process. Process acts as a thread.
type Process struct {
	sync.RWMutex

	ctx context.Context
	cfn context.CancelFunc

	stdin  io.Writer
	stdout io.Reader
	stderr io.Reader

	command       *exec.Cmd
	logger        *zap.SugaredLogger
	waitDone      chan error
	waitListeners []chan error
}

// NewProcess returns a new Process instance.
func NewProcess(logger *zap.SugaredLogger) *Process {
	return &Process{
		logger: logger,
	}
}

// Start starts the bedrock
func (p *Process) Start() error {
	var err error

	p.Lock()

	if p.command != nil || p.ctx != nil || p.cfn != nil {
		p.Unlock()
		return errors.New("bedrock process appears to already be started")
	}

	executable, err := os.Executable()
	if err != nil {
		p.Unlock()
		return errors.New("unable to get current executable location")
	}

	dir := path.Dir(executable)

	p.ctx, p.cfn = context.WithCancel(context.Background())
	p.command = exec.Command(fmt.Sprintf("%s/bedrock_server", dir))
	p.command.Dir = dir
	p.command.Env = []string{"LD_LIBRARY_PATH=."}

	// Start the command within it's own process group, so that it doesn't receive the signals
	// that this application receives. This allows us to correctly quit the server.
	p.command.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	p.stdin, err = p.command.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "failed to get stdin pipe")
	}

	p.stdout, err = p.command.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to get stdout pipe")
	}

	p.stderr, err = p.command.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "failed to get stderr pipe")
	}

	p.pipeOutput(p.stdout, "stdout")
	p.pipeOutput(p.stderr, "stderr")

	err = p.command.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start bedrock server")
	}

	p.Unlock()

	p.logger.Info("started bedrock server thread")

	return p.command.Wait()
}

// pipeOutput reads lines from one of the process's output pipes and outputs them using this
// application's logger, so it can be the same format as our application's logs.
func (p *Process) pipeOutput(pr io.Reader, name string) {
	go func() {
		reader := bufio.NewReader(pr)

		for {
			// Check if we should stop.
			if p.ctx.Err() != nil {
				return
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				// TODO: Something.
				return
			}

			p.logger.Infow(strings.TrimSpace(line), "pipe", name)
		}
	}()
}

// Stop doesn't actually stop this Thread, it will wait for it to stop, and eventually kill it if
// it hasn't stopped already. It doesn't actually stop the Thread because the child process spawned
// by this Thread will receive the same signal
func (p *Process) Stop() error {
	// First, attempt to quit the server.
	_, err := io.WriteString(p.stdin, "stop\n")
	if err != nil {
		return errors.Wrap(err, "failed to send 'stop' to bedrock server")
	}

	return nil
}

// Kill attempts to kill the bedrock server process - to try clean up as best we can.
func (p *Process) Kill() error {
	if p.command == nil || p.command.Process == nil {
		return nil
	}

	return p.command.Process.Kill()
}
