// Package compiler provides a compiler for the goja runtime.
package compiler

import (
	"context"
	"fmt"

	"github.com/Mzack9999/goja"
	"github.com/kitabisa/go-ci"

	"github.com/projectdiscovery/nuclei/v3/pkg/protocols/common/generators"
	"github.com/projectdiscovery/nuclei/v3/pkg/types"
	contextutil "github.com/projectdiscovery/utils/context"
	"github.com/projectdiscovery/utils/errkit"
	stringsutil "github.com/projectdiscovery/utils/strings"
)

var (
	// ErrJSExecDeadline is the error returned when allotted time for script execution exceeds
	ErrJSExecDeadline = errkit.New("js engine execution deadline exceeded").SetKind(errkit.ErrKindDeadline).Build()
)

// Compiler provides a runtime to execute goja runtime
// based javascript scripts efficiently while also
// providing them access to custom modules defined in libs/.
type Compiler struct{}

// New creates a new compiler for the goja runtime.
func New() *Compiler {
	return &Compiler{}
}

// ExecuteOptions provides options for executing a script.
type ExecuteOptions struct {
	// ExecutionId is the id of the execution
	ExecutionId string

	// Callback can be used to register new runtime helper functions
	// ex: export etc
	Callback func(runtime *goja.Runtime) error

	// Cleanup is extra cleanup function to be called after execution
	Cleanup func(runtime *goja.Runtime)

	// Source is original source of the script
	Source *string

	Context context.Context

	TimeoutVariants *types.Timeouts

	// Manually exported objects
	exports map[string]any
}

// ExecuteArgs is the arguments to pass to the script.
type ExecuteArgs struct {
	Args        map[string]any //these are protocol variables
	TemplateCtx map[string]any // templateCtx contains template scoped variables
}

// Map returns a merged map of the TemplateCtx and Args fields.
func (e *ExecuteArgs) Map() map[string]any {
	return generators.MergeMaps(e.TemplateCtx, e.Args)
}

// NewExecuteArgs returns a new execute arguments.
func NewExecuteArgs() *ExecuteArgs {
	return &ExecuteArgs{
		Args:        make(map[string]any),
		TemplateCtx: make(map[string]any),
	}
}

// ExecuteResult is the result of executing a script.
type ExecuteResult map[string]any

// Map returns the map representation of the ExecuteResult
func (e ExecuteResult) Map() map[string]any {
	if e == nil {
		return make(map[string]any)
	}
	return e
}

// NewExecuteResult returns a new execute result instance
func NewExecuteResult() ExecuteResult {
	return make(map[string]any)
}

// GetSuccess returns whether the script was successful or not.
func (e ExecuteResult) GetSuccess() bool {
	if e == nil {
		return false
	}
	val, ok := e["success"].(bool)
	if !ok {
		return false
	}
	return val
}

// ExecuteWithOptions executes a script with the provided options.
func (c *Compiler) ExecuteWithOptions(program *goja.Program, args *ExecuteArgs, opts *ExecuteOptions) (ExecuteResult, error) {
	if opts == nil {
		opts = &ExecuteOptions{Context: context.Background()}
	}
	if args == nil {
		args = NewExecuteArgs()
	}
	// handle nil maps
	if args.TemplateCtx == nil {
		args.TemplateCtx = make(map[string]any)
	}
	if args.Args == nil {
		args.Args = make(map[string]any)
	}
	// merge all args into templatectx
	args.TemplateCtx = generators.MergeMaps(args.TemplateCtx, args.Args)

	// execute with context and timeout

	ctx, cancel := context.WithTimeoutCause(opts.Context, opts.TimeoutVariants.JsCompilerExecutionTimeout, ErrJSExecDeadline)
	defer cancel()
	// execute the script
	results, err := contextutil.ExecFuncWithTwoReturns(ctx, func() (val goja.Value, err error) {
		// TODO(dwisiswant0): remove this once we get the RCA.
		defer func() {
			if ci.IsCI() {
				return
			}

			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()

		return ExecuteProgram(program, args, opts)
	})
	if err != nil {
		if val, ok := err.(*goja.Exception); ok {
			if x := val.Unwrap(); x != nil {
				err = x
			}
		}
		e := NewExecuteResult()
		e["error"] = err.Error()
		return e, err
	}
	var res ExecuteResult
	if opts.exports != nil {
		res = ExecuteResult(opts.exports)
		opts.exports = nil
	} else {
		res = NewExecuteResult()
	}
	res["response"] = results.Export()
	res["success"] = results.ToBoolean()
	return res, nil
}

// if the script uses export/ExportAS tokens then we can run it in IIFE mode
// but if not we can't run it
func CanRunAsIIFE(script string) bool {
	return stringsutil.ContainsAny(script, exportAsToken, exportToken)
}

// SourceIIFEMode is a mode where the script is wrapped in a function and compiled.
// This is used when the script is not exported or exported as a function.
func SourceIIFEMode(script string, strict bool) (*goja.Program, error) {
	val := fmt.Sprintf(`
		(function() {
			%s
		})()
	`, script)
	return goja.Compile("", val, strict)
}

// SourceAutoMode is a mode where the script is wrapped in a function and compiled.
// This is used when the script is exported or exported as a function.
func SourceAutoMode(script string, strict bool) (*goja.Program, error) {
	if !CanRunAsIIFE(script) {
		// this will not be run in a pooled runtime
		return goja.Compile("", script, strict)
	}
	val := fmt.Sprintf(`
		(function() {
			%s
		})()
	`, script)
	return goja.Compile("", val, strict)
}
