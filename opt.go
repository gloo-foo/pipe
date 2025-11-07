package command

type flags struct {
	pipeFail pipeFailFlag
	buffered bufferedFlag
	verbose  verboseFlag
	dryRun   dryRunFlag
	maxProcs MaxProcs
}

type pipeFailFlag bool

const (
	PipeFail   pipeFailFlag = true
	NoPipeFail pipeFailFlag = false
)

type bufferedFlag bool

const (
	Buffered   bufferedFlag = true
	Unbuffered bufferedFlag = false
)

type verboseFlag bool

const (
	Verbose verboseFlag = true
	Quiet   verboseFlag = false
)

type dryRunFlag bool

const (
	DryRun   dryRunFlag = true
	NoDryRun dryRunFlag = false
)

type MaxProcs int

func (f pipeFailFlag) Configure(flags *flags) { flags.pipeFail = f }
func (f bufferedFlag) Configure(flags *flags) { flags.buffered = f }
func (f verboseFlag) Configure(flags *flags)  { flags.verbose = f }
func (f dryRunFlag) Configure(flags *flags)   { flags.dryRun = f }
func (m MaxProcs) Configure(flags *flags)     { flags.maxProcs = m }
