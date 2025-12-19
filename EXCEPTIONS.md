# Refactoring Exceptions

Documenting files where functions were intentionally left in place during the cmd/ cleanup.

## cloud/events/events.go

Functions left in place:
- `type cloudEvent` - Local type specific to cloud WebSocket events
- `func (e *cloudEvent) getDeviceID` - Method on local type
- `func readEvents` - Tightly coupled to WebSocket event loop
- `func handleMessage` - Uses package-level flag variables
- `func displayEvent` - Uses local cloudEvent type
- `func printIndentedJSON` - Simple helper for displayEvent
- `func handleReadError` - WebSocket-specific error handling
- `func isExpectedClosure` - WebSocket-specific error checking

**Rationale**: These functions are tightly coupled to:
1. The local `cloudEvent` type (not reusable elsewhere)
2. Package-level flag variables (`rawFlag`, `formatFlag`, `filterDeviceFlag`, `filterEventFlag`)
3. The WebSocket event loop lifecycle

The `buildWebSocketURL` function was extracted to `shelly.BuildCloudWebSocketURL` as it had no local dependencies.
