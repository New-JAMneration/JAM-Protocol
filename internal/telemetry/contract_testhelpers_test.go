package telemetry

// clearRegistry drops every registered BridgeContract. Test-only — the
// production binary should never call this. Lives in a *_test.go file
// so the production build excludes it entirely; production code in the
// telemetry package cannot reach it.
func clearRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = nil
}
