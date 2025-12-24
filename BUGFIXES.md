# Critical Bug Fixes

This document details the critical bug fixes implemented based on code review feedback.

## 🔴 Critical: Goroutine Leaks Fixed

**Problem:** Workflow cancellation didn't stop triggers. Hot-reload or workflow removal would leave old cron/filewatch goroutines running, causing double-firing and resource leaks.

**Root Cause:**
- `cmd/agent.go` created a context but triggers never accepted or observed it
- No stop mechanism in `internal/trigger/cron.go` and `internal/trigger/filewatch.go`
- Cron schedulers and file watchers ran forever

**Solution:**
1. Updated trigger signatures to accept `context.Context`:
   ```go
   func StartCronTrigger(ctx context.Context, wf *workflow.Workflow) error
   func StartFileWatchTrigger(ctx context.Context, wf *workflow.Workflow) error
   ```

2. Added context cancellation watchers:
   ```go
   go func() {
       <-ctx.Done()
       // Cleanup: stop cron, close file watcher, unregister workflow
   }()
   ```

3. Updated all callers to pass context

**Impact:** ✅ No more goroutine leaks, proper cleanup on hot-reload/shutdown

---

## 🟠 High: YAML Field Name Mismatch Fixed

**Problem:** HTTP expectations in shipped workflows weren't applied. Struct tags used camelCase while YAML used snake_case, causing silent validation failures.

**Example:**
```yaml
# workflows/api-health-check.yaml
expect_status: [200]           # ❌ Silently ignored!
expect_body_contains: "healthy" # ❌ Silently ignored!
```

```go
// internal/workflow/types.go (before fix)
ExpectStatus       interface{} `yaml:"expectStatus,omitempty"` // ❌ Wrong tag!
ExpectBodyContains string      `yaml:"expectBodyContains,omitempty"` // ❌ Wrong tag!
```

**Solution:**
Updated struct tags to match YAML convention:
```go
ExpectStatus       interface{} `yaml:"expect_status,omitempty" json:"expectStatus,omitempty"`
ExpectBodyContains string      `yaml:"expect_body_contains,omitempty" json:"expectBodyContains,omitempty"`
```

**Impact:** ✅ HTTP validations now work correctly, responses are properly validated

---

## 🟡 Medium: Type Conversion for expect_status Fixed

**Problem:** YAML unmarshals lists into `[]interface{}` but code expected `[]int`. Common configs like `expect_status: [200, 204]` would fail validation.

**Root Cause:** YAML numeric values unmarshal as `float64`, not `int`

**Solution:**
Added comprehensive type handling in `internal/action/http.go`:
```go
// Handle multiple input types
switch v := action.ExpectStatus.(type) {
case int:
    expectedStatuses = append(expectedStatuses, v)
case float64:
    // YAML numbers become float64
    expectedStatuses = append(expectedStatuses, int(v))
case []interface{}:
    // Convert each item
    for i, item := range v {
        switch code := item.(type) {
        case int:
            expectedStatuses = append(expectedStatuses, code)
        case float64:
            expectedStatuses = append(expectedStatuses, int(code))
        default:
            return fmt.Errorf("invalid status code type at index %d: %T", i, code)
        }
    }
}
```

**Impact:** ✅ All status code formats work: single int, list of ints, YAML floats

---

## 🟢 Low: Filewatch Event Validation Added

**Problem:** Invalid filewatch event names weren't validated at parse time. Only logged errors at runtime, leading to workflows that never fire.

**Example:**
```yaml
trigger:
  type: filewatch
  path: /tmp/test
  events:
    - create
    - modified  # ❌ Invalid! Should be "write"
```

**Solution:**
Added parse-time validation in `internal/parser/parser.go`:
```go
func validateFileWatchEvents(events []string) error {
    validEvents := map[string]bool{
        "create": true,
        "write":  true,  // ← Correct name
        "remove": true,
        "rename": true,
        "chmod":  true,
    }

    for _, event := range events {
        if !validEvents[event] {
            return fmt.Errorf("invalid filewatch event: '%s'. Valid events are: create, write, remove, rename, chmod", event)
        }
    }
    return nil
}
```

**Impact:** ✅ Fail fast with clear error message during workflow validation

---

## Files Changed

### Modified Files
- `internal/trigger/cron.go` - Context-aware shutdown
- `internal/trigger/filewatch.go` - Context-aware shutdown
- `cmd/agent.go` - Pass context to triggers
- `cmd/run.go` - Pass context to triggers, add context import
- `internal/workflow/types.go` - Fixed YAML struct tags
- `internal/action/http.go` - Type conversion for expect_status
- `internal/parser/parser.go` - Event validation, removed strict type check

### Tests Verified
- ✅ `./autozap validate workflows/api-health-check.yaml` - expect_status works
- ✅ `./autozap validate workflows/file-monitor.yaml` - event validation works
- ✅ Build succeeds with no errors
- ✅ All workflow validations pass

## Before vs After

### Before
- ❌ Goroutine leaks on hot-reload
- ❌ HTTP validations silently ignored
- ❌ expect_status: [200, 204] fails
- ❌ Invalid events discovered only at runtime

### After
- ✅ Proper cleanup on context cancellation
- ✅ HTTP validations work correctly
- ✅ All status code formats supported
- ✅ Invalid events caught at parse time

## Testing Recommendations

1. **Test Hot-Reload:**
   ```bash
   ./autozapctl start
   # Add/remove workflows while running
   # No goroutine leaks!
   ```

2. **Test HTTP Validations:**
   ```yaml
   actions:
     - type: http
       url: "https://api.example.com"
       expect_status: [200, 201, 204]  # Now works!
       expect_body_contains: "success"  # Now works!
   ```

3. **Test Event Validation:**
   ```bash
   # This will fail at parse time (not runtime)
   ./autozap validate bad-workflow.yaml
   # Error: invalid filewatch event: 'modified'
   ```

## Credits

All fixes based on excellent code review feedback identifying:
- Critical goroutine leaks
- Silent validation failures
- Type conversion issues
- Missing parse-time validation
