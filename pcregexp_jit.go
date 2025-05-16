package pcregexp

// Default stack sizes for JIT compilation
// These values are used to set the initial and maximum stack sizes for the JIT
// compiler when compiling regular expressions.
// The default values are set to reasonable sizes for most use cases.
// The initial size is the amount of stack space allocated when the JIT compiler
// is first created, and the maximum size is the maximum amount  of stack space
// that can be allocated during JIT compilation. These values can be adjusted
// based on the specific needs of the application and the complexity of the
// regular expressions being compiled by using the [SetJITStackSize] function.
const (
	// DefaultJITStackStartSize is the default initial size of the JIT stack.
	DefaultJITStackStartSize = uint64(32 * 1024) // 32 KiB initial size

	// DefaultJITStackMaxSize is the default maximum size of the JIT stack.
	DefaultJITStackMaxSize = uint64(512 * 1024) // 512 KiB max size
)

var (
	// Default JIT option and stack sizes for internal use
	defaultJITOption         = JITComplete
	defaultJITStackStartSize = DefaultJITStackStartSize
	defaultJITStackMaxSize   = DefaultJITStackMaxSize
)

// JITOption represents the JIT compilation options.
// It is used to specify the level of JIT compilation to be performed.
//
// The default option is [JITComplete], which enables full JIT compilation.
// The JIT compilation options are used to control the behavior of the JIT
// compiler when compiling regular expressions. The options can be set using
// the [SetJITOption] function.
type JITOption uint32

const (
	JITNoJit JITOption = iota
	JITComplete
	JITPartialSoft
	_
	JITPartialHard // 4
)

// SetJITOption sets the default JIT option used for JIT compilation.
// The option parameter specifies the JIT option to be used.
// The default option is [JITComplete], which enables full JIT compilation.
func SetJITOption(option JITOption) {
	defaultJITOption = option
}

// SetJITStackSize sets the default JIT stack sizes used for JIT compilation.
// The startSize parameter sets the initial stack size, and maxSize sets
// the maximum size it can grow to.
//
// The default number is 32 KiB for startSize and 512 KiB for maxSize.
func SetJITStackSize(startSize, maxSize uint64) {
	defaultJITStackStartSize = startSize
	defaultJITStackMaxSize = maxSize
}
