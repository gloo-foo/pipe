package command

type pipeFailFlag bool

const (
	PipeFail   pipeFailFlag = true
	PipeNoFail pipeFailFlag = false
)

type pipeBufferedFlag bool

const (
	PipeBuffered   pipeBufferedFlag = true
	PipeUnbuffered pipeBufferedFlag = false
)

type pipeVerboseFlag bool

const (
	PipeVerbose pipeVerboseFlag = true
	PipeQuiet   pipeVerboseFlag = false
)

type pipeDryRunFlag bool

const (
	PipeDryRun   pipeDryRunFlag = true
	PipeNoDryRun pipeDryRunFlag = false
)

type PipeMaxProcs int

type flags struct {
	pipeFail pipeFailFlag
	buffered pipeBufferedFlag
	verbose  pipeVerboseFlag
	dryRun   pipeDryRunFlag
	maxProcs PipeMaxProcs
}

func (f pipeFailFlag) Configure(flags *flags)     { flags.pipeFail = f }
func (f pipeBufferedFlag) Configure(flags *flags) { flags.buffered = f }
func (f pipeVerboseFlag) Configure(flags *flags)  { flags.verbose = f }
func (f pipeDryRunFlag) Configure(flags *flags)   { flags.dryRun = f }
func (m PipeMaxProcs) Configure(flags *flags)     { flags.maxProcs = m }
