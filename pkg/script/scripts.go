package script

import _ "embed"

// ConsumeKeyOnceAtomic deletes the key and returns 1 when it exists, otherwise returns 0.
//
//go:embed consume_key_once_atomic.lua
var ConsumeKeyOnceAtomic string
