# TODO

* [x] Implement PCRE2 JIT compilation support
  * Use native PCRE2 API JIT functions for improved performance
  * Add JIT compilation options and configurations
  * Implement memory management for JIT-compiled patterns
* [ ] Implement these methods (**std `regexp` compatibility**):
  * [ ] `NumSubexp`
  * [ ] `LiteralPrefix`
  * [ ] `Longest`
  * [ ] `SubexpNames`
  * [ ] `SubexpIndex`
* [ ] Add these methods:
  * [ ] `ReplaceAllWithSubstitute` (`pcre2_substitute`)
  * [ ] `PatternInfo` (`pcre2_pattern_info`)
* [ ] Add these functions:
  * [ ] `GetErrorMessage` (`pcre2_get_error_message`)
* [ ] Support these match context fields:
  * [ ] `OffsetLimit`
  * [ ] `HeapLimit`
  * [x] `MatchLimit`
  * [x] ~~`RecursionLimit`~~ _(Obsolete)_ => `DepthLimit`
* [ ] Implement iterator for global find ops (`pcre2_next_match` since PCRE2 10.46)
  * [ ] `FindIter`
  * [ ] `Next`
  * [ ] `Group`