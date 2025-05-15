package pcregexp

import "fmt"

// MatchContext provides configuration for regex matching operations.
type MatchContext struct {
	// TODO(dwisiswant0): Add support for this field.
	// // OffsetLimit is the maximum offset in the subject string.
	// OffsetLimit uint64

	// TODO(dwisiswant0): Add support for this field.
	// // HeapLimit is the maximum heap memory in kibibytes (KiB).
	// HeapLimit uint32

	// MatchLimit is the maximum number of matches to allow.
	MatchLimit uint32

	// DepthLimit is the maximum recursion depth to allow.
	DepthLimit uint32

	// ptr is the internal match context pointer.
	ptr uintptr
}

// Default match context used by all regex operations unless overridden
var defaultMatchCtx *MatchContext

// SetMatchContext sets the global match context used by all regex operations.
//
// This is useful to globally limit regex matching complexity to prevent ReDoS
// attacks. If the context is not set, no limits are enforced.
func SetMatchContext(ctx MatchContext) error {
	if defaultMatchCtx != nil {
		// Free the old context
		pcre2_match_context_free(defaultMatchCtx.ptr)
		defaultMatchCtx = nil
	}

	if ctx.MatchLimit == 0 && ctx.DepthLimit == 0 {
		return nil
	}

	ctx.ptr = pcre2_match_context_create(0)
	if ctx.ptr == 0 {
		return fmt.Errorf("could not create match context")
	}

	if ctx.MatchLimit > 0 {
		if result := pcre2_set_match_limit(ctx.ptr, ctx.MatchLimit); result != 0 {
			pcre2_match_context_free(ctx.ptr)
			return fmt.Errorf("could not set match limit, error code: %d", result)
		}
	}

	if ctx.DepthLimit > 0 {
		if result := pcre2_set_depth_limit(ctx.ptr, ctx.DepthLimit); result != 0 {
			pcre2_match_context_free(ctx.ptr)
			return fmt.Errorf("could not set recursion limit, error code: %d", result)
		}
	}

	defaultMatchCtx = &ctx
	return nil
}
