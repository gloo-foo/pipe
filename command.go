package command

import (
	`context`
	`errors`
	`fmt`
	`io`
	`log/slog`
	`sync`

	yup `github.com/gloo-foo/framework`
)

type command Inputs[flags]

func Pipeline(parameters ...any) yup.Command {
	return command(args[flags](parameters...))
}

func (p command) Executor() yup.CommandExecutor {
	return func(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
		// Use the commands from the pipeline
		commands := p.commands

		if len(commands) == 0 {
			return fmt.Errorf("no commands in pipeline")
		}

		// Single command - no piping needed
		if len(commands) == 1 {
			executor := commands[0].Executor()
			return executor(ctx, stdin, stdout, stderr)
		}

		// Multiple commands - set up pipeline
		type cmdError struct {
			index int
			err   error
		}

		// Create a cancellable context for all commands
		pipelineCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		var (
			wg    sync.WaitGroup
			errMu sync.Mutex
			errs  []cmdError
		)

		// Create pipes to connect commands
		pipes := make([]*io.PipeReader, len(commands)-1)
		pipeWriters := make([]*io.PipeWriter, len(commands)-1)

		for i := 0; i < len(commands)-1; i++ {
			pipes[i], pipeWriters[i] = io.Pipe()
		}

		// Launch each command in a goroutine
		for i, cmd := range commands {
			wg.Add(1)

			// Determine input for this command
			var cmdStdin io.Reader
			var pipeReader *io.PipeReader
			if i == 0 {
				cmdStdin = stdin
			} else {
				pipeReader = pipes[i-1]
				cmdStdin = pipeReader
			}

			// Determine output for this command
			var cmdStdout io.Writer
			if i == len(commands)-1 {
				cmdStdout = stdout
			} else {
				cmdStdout = pipeWriters[i]
			}

			go func(index int, command yup.Command, in io.Reader, out io.Writer, pipeIn *io.PipeReader) {
				defer wg.Done()

				// When the last command finishes, cancel the pipeline context
				// This signals all upstream commands to stop
				if index == len(commands)-1 {
					defer func() {
						cancel()
						// Also close all upstream pipe readers to unblock any blocked writes
						for j := 0; j < index; j++ {
							pipes[j].Close()
						}
					}()
				}

				// Close the pipe reader when done (if not the first command)
				// This signals upstream that we're done reading and unblocks any Write()
				if pipeIn != nil {
					defer pipeIn.Close()
				}

				// Close the pipe writer when done (if not the last command)
				if index < len(commands)-1 {
					defer func(writer *io.PipeWriter) {
						err := writer.Close()
						if err != nil {
							panic(err)
						}
					}(pipeWriters[index])
				}

				executor := command.Executor()
				err := executor(pipelineCtx, in, out, stderr)

				if err != nil {
					// For non-last commands, ignore context cancellation and pipe-related errors
					// This is expected when downstream commands finish early (e.g., yes | head)
					if index < len(commands)-1 {
						// Context was cancelled (downstream finished)
						if errors.Is(err, context.Canceled) {
							return
						}
						// Check for broken pipe errors
						if errors.Is(err, io.ErrClosedPipe) || errors.Is(err, io.EOF) {
							return
						}
						// Also check the error string as a fallback
						errStr := err.Error()
						if errStr == "io: read/write on closed pipe" || errStr == "write on closed pipe" {
							return
						}
					}

					errMu.Lock()
					errs = append(errs, cmdError{index: index, err: err})
					errMu.Unlock()

					// If pipefail is enabled, close downstream pipes to signal error
					if bool(p.Flags.pipeFail) {
						for j := index; j < len(pipeWriters); j++ {
							err := pipeWriters[j].CloseWithError(err)
							if err != nil {
								panic(err)
							}
						}
					}
				}
			}(i, cmd, cmdStdin, cmdStdout, pipeReader)
		}

		// Wait for all commands to complete
		wg.Wait()

		// Handle errs based on pipefail flag
		if len(errs) > 0 {
			if bool(p.Flags.pipeFail) {
				// Return first error if pipefail is enabled
				return fmt.Errorf("command %d: %w", errs[0].index, errs[0].err)
			}
			// If pipefail is disabled, only return error from last command
			lastCommandIndex := len(commands) - 1
			for _, cmdErr := range errs {
				if cmdErr.index == lastCommandIndex {
					return fmt.Errorf("command %d: %w", cmdErr.index, cmdErr.err)
				}
			}
		}

		return nil
	}
}

type Inputs[O any] struct {
	Flags    O
	commands []yup.Command
}

func args[O any](parameters ...any) (result Inputs[O]) {
	var (
		options []yup.Switch[O]
	)
	for _, arg := range parameters {
		switch v := arg.(type) {
		case yup.Switch[O]:
			options = append(options, v)
		case yup.Command:
			result.commands = append(result.commands, v)
		default:
			slog.Warn("Unknown argument type", "arg", v, "type", fmt.Sprintf("%T/%T", arg, v))
		}
	}
	result.Flags = configure(options...)
	return result
}

func configure[T any](opts ...yup.Switch[T]) T {
	def := new(T)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt.Configure(def)
	}
	return *def
}
