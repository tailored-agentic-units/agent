// Package registry manages named agent configurations with lazy instantiation.
// Configs are stored at registration time; agents are created on first Get call.
// Thread-safe for concurrent access.
//
// The registry resolves provider and format dependencies from their respective
// global registries (provider.Create and format.Create) at agent creation time.
//
// # Usage
//
//	r := registry.NewRegistry()
//
//	err := r.Register("assistant", agentConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	a, err := r.Get("assistant")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	resp, err := a.Chat(ctx, messages)
package registry
