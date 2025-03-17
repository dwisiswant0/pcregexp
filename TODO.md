# TODO

* [ ] Implement PCRE2 JIT compilation support
  * Use native PCRE2 API JIT functions for improved performance
  * Add JIT compilation options and configurations
  * Implement memory management for JIT-compiled patterns
* [ ] Implement these methods:
  * [ ] `NumSubexp`
  * [ ] `LiteralPrefix`
  * [ ] `Longest`
  * [ ] `SubexpNames`
  * [ ] `SubexpIndex`
* [ ] Support these match context fields:
  * [ ] `OffsetLimit`
  * [ ] `HeapLimit`
  * [x] `MatchLimit`
  * [x] `RecursionLimit`