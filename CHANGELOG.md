# Changelog

## [v0.1.1] - 2026-04-06

### Fixed

- Fix nil pointer dereference in streaming responses when format returns `(nil, nil)` for unrecognized stream events (e.g., Converse format skipping unknown event types)

## [v0.1.0] - 2026-04-06

### Added

- Agent interface with Chat, ChatStream, Vision, VisionStream, Tools, and Embed methods
- Explicit dependency injection constructor: `New(cfg, provider, format)`
- Client package with HTTP orchestration, retry logic, and health tracking
- Request package with ChatRequest, VisionRequest, ToolsRequest, and EmbeddingsRequest
- Registry package for named agent configuration with lazy instantiation
- Mock package with MockAgent, MockClient, MockProvider, and MockFormat
- AgentError type with categorization (init, llm) and contextual metadata
