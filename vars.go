package pcregexp

// matchFunc is a function type that defines the signature for the PCRE2 match
// function.
//
// The arguments for pcre2_jit_match() are exactly the same as for pcre2_match().
// Ref: https://pcre2project.github.io/pcre2/doc/pcre2_jit_match/
type matchFunc func(code uintptr, subject *uint8, length uint64, startoffset uint64, options uint32, matchData uintptr, matchContext uintptr) int32

var (
	// globalFinalizerObject is used to attach a finalizer for cleanup
	globalFinalizerObject = new(int)

	// pcre2_compile_8 signature:
	//   pcre2_code *pcre2_compile_8(PCRE2_SPTR pattern, PCRE2_SIZE length,
	//       uint32_t options, int *errorcode, PCRE2_SIZE *erroroffset,
	//       pcre2_compile_context *ccontext);
	pcre2_compile func(pattern *uint8, length uint64, options uint32, errorcode *int32, erroroffset *uint64, compileContext uintptr) uintptr

	// pcre2_code_free_8: void pcre2_code_free_8(pcre2_code *code);
	pcre2_code_free func(code uintptr)

	// pcre2_pattern_info_8: int pcre2_pattern_info_8(const pcre2_code *code,
	//    uint32_t what, void *where);
	pcre2_pattern_info func(code uintptr, what uint32, where uintptr) int32

	// pcre2_match_8: int pcre2_match_8(const pcre2_code *code,
	//    PCRE2_SPTR subject, PCRE2_SIZE length, PCRE2_SIZE startoffset,
	//	  uint32_t options, pcre2_match_data *match_data,
	// 	  pcre2_match_context *mcontext);
	pcre2_match matchFunc

	// pcre2_match_data_create_from_pattern_8:
	// 	  pcre2_match_data *pcre2_match_data_create_from_pattern_8(
	// 	  	  const pcre2_code *code, pcre2_general_context *gcontext);
	pcre2_match_data_create_from_pattern func(code uintptr, generalContext uintptr) uintptr

	// pcre2_match_data_free_8:
	// 	  void pcre2_match_data_free_8(pcre2_match_data *match_data);
	pcre2_match_data_free func(matchData uintptr)

	// pcre2_get_ovector_pointer_8:
	// 	  PCRE2_SIZE *pcre2_get_ovector_pointer_8(pcre2_match_data *match_data);
	pcre2_get_ovector_pointer func(matchData uintptr) *uint64

	// Match context functions for timeout support
	// pcre2_match_context_create_8:
	//    pcre2_match_context *pcre2_match_context_create_8(
	//        pcre2_general_context *gcontext);
	pcre2_match_context_create func(generalContext uintptr) uintptr

	// pcre2_match_context_free_8:
	//    void pcre2_match_context_free_8(pcre2_match_context *mcontext);
	pcre2_match_context_free func(matchContext uintptr)

	// // pcre2_set_offset_limit_8:
	// //    int pcre2_set_offset_limit_8(pcre2_match_context *mcontext,
	// //        PCRE2_SIZE value);
	// pcre2_set_offset_limit func(matchContext uintptr, value uint64) int32

	// // pcre2_set_heap_limit_8:
	// //    int pcre2_set_heap_limit_8(pcre2_match_context *mcontext,
	// //        uint32_t value);
	// pcre2_set_heap_limit func(matchContext uintptr, value uint32) int32

	// pcre2_set_match_limit_8:
	//    int pcre2_set_match_limit_8(pcre2_match_context *mcontext,
	//        uint32_t value);
	pcre2_set_match_limit func(matchContext uintptr, value uint32) int32

	// // pcre2_set_recursion_limit_8:
	// //    int pcre2_set_recursion_limit_8(pcre2_match_context *mcontext,
	// //    uint32_t value);
	// //
	// // NOTE(dwisiswant0): This function is became obsolete at PCRE2 10.30.
	// // See: https://pcre2project.github.io/pcre2/doc/pcre2api/#:~:text=PCRE2%20NATIVE%20API%20OBSOLETE%20FUNCTIONS
	// pcre2_set_recursion_limit func(matchContext uintptr, value uint32) int32

	// pcre2_set_depth_limit_8:
	//    int pcre2_set_depth_limit_8(pcre2_match_context *mcontext,
	//        uint32_t value);
	pcre2_set_depth_limit func(matchContext uintptr, value uint32) int32

	// pcre2_jit_compile:
	//    int pcre2_jit_compile_8(pcre2_code *code, uint32_t options);
	pcre2_jit_compile func(code uintptr, options uint32) int32

	// pcre2_match_8: int pcre2_match_8(const pcre2_code *code,
	//    PCRE2_SPTR subject, PCRE2_SIZE length, PCRE2_SIZE startoffset,
	//	  uint32_t options, pcre2_match_data *match_data,
	// 	  pcre2_match_context *mcontext);
	pcre2_jit_match matchFunc

	// pcre2_jit_stack_create:
	//    pcre2_jit_stack *pcre2_jit_stack_create_8(PCRE2_SIZE startsize,
	//        PCRE2_SIZE maxsize, pcre2_general_context *gcontext);
	pcre2_jit_stack_create func(startsize uint64, maxsize uint64, generalContext uintptr) uintptr

	// pcre2_jit_stack_free:
	//    void pcre2_jit_stack_free_8(pcre2_jit_stack *stack);
	pcre2_jit_stack_free func(stack uintptr)

	// pcre2_jit_stack_assign:
	//    int pcre2_jit_stack_assign_8(pcre2_match_context *mcontext,
	//        pcre2_jit_callback callback, void *callback_data)
	pcre2_jit_stack_assign func(matchContext uintptr, callback uintptr, data uintptr) int32
)
