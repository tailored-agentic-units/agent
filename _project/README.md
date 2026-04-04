# agent

Stateless LLM agent composition. Wires protocol types, wire formats, and transport providers together through HTTP client and request infrastructure.

## Vision

A clean agent abstraction where the Agent is stateless transport — it takes exactly the messages you give it and sends them through the format layer via the provider's transport. Callers own context management, enabling explicit strategies (sliding window, summarization, priority-based retention, session branching). The composition pattern: callers resolve a provider and format, inject them into the agent constructor, and interact through the `[]Message`-based interface.

## Core Premise

The Agent interface aligns with the raw LLM API contract: message arrays in, structured responses out. No hidden message construction, no internal state. The agent composes three independent concerns — protocol types (tau/protocol), wire serialization (tau/format), and transport (tau/provider) — through explicit dependency injection.

## Phases

| Phase | Focus Area | Version Target |
|-------|-----------|----------------|
| Phase 1 - Foundation | Request/client infrastructure, Agent interface and implementation, mock and registry | v0.1.0 |

## Architecture

```
agent (root)     — Agent interface, New(cfg, provider, format), implementation
  request/       — Request interface, Chat/Vision/Tools/Embeddings request types
  client/        — Client interface, HTTP execution, retry, health tracking
  mock/          — MockAgent, MockClient, MockProvider, MockFormat
  registry/      — AgentRegistry (named agents, lazy instantiation)
```

### Dependency Hierarchy

```
Level 0: request/ (imports tau/protocol, tau/format, tau/provider)
Level 1: client/ (imports request/, tau/protocol, tau/format, tau/provider)
Level 2: agent root (imports client/, tau/protocol, tau/format, tau/provider)
Level 3: mock/ (imports agent, client, tau/format, tau/provider, tau/protocol)
Level 4: registry/ (imports agent root, tau/protocol/config)
```

## Source References

### Root Package (Agent Interface + Implementation)

**REWRITE** — combines kernel's Agent interface structure with go-agents' three-layer composition.

**Kernel source**: `~/tau/kernel/agent/agent.go` (323 lines), `~/tau/kernel/agent/errors.go` (163 lines)
**go-agents source**: `~/code/go-agents/pkg/agent/agent.go` (419 lines), `~/code/go-agents/pkg/agent/errors.go` (155 lines)

**Deviations from kernel Agent**:

Kernel Agent interface (`~/tau/kernel/agent/agent.go`):
```go
type Agent interface {
    ID() string
    Client() client.Client
    Provider() providers.Provider
    Model() *model.Model
    Chat(ctx context.Context, prompt []protocol.Message, opts ...map[string]any) (*response.ChatResponse, error)
    ChatStream(ctx context.Context, prompt []protocol.Message, opts ...map[string]any) (<-chan *response.StreamingChunk, error)
    // ... Vision, Tools, ToolsStream, Embed, Audio
}
```

tau/agent Agent interface (new):
```go
type Agent interface {
    ID() string
    Client() client.Client
    Provider() provider.Provider    // singular "provider" module, not "providers"
    Format() format.Format          // NEW — go-agents addition
    Model() *model.Model
    Chat(ctx context.Context, messages []protocol.Message, opts ...map[string]any) (*response.Response, error)           // unified Response
    ChatStream(ctx context.Context, messages []protocol.Message, opts ...map[string]any) (<-chan *response.StreamingResponse, error)  // unified StreamingResponse
    // ... Vision, Tools with same signature pattern, ToolDefinition instead of Tool
    Embed(ctx context.Context, input string, opts ...map[string]any) (*response.EmbeddingsResponse, error)
}
```

Changes:
- `*response.ChatResponse` → `*response.Response` (unified)
- `*response.ToolsResponse` → `*response.Response` (unified)
- `<-chan *response.StreamingChunk` → `<-chan *response.StreamingResponse` (unified)
- `[]protocol.Tool` → `[]format.ToolDefinition` (moved to format layer)
- `Format() format.Format` accessor added
- `Audio()` method removed (deferred)
- `images []string` → `images []format.Image` (structured type)

**Deviations from go-agents Agent**:

go-agents Agent interface (`~/code/go-agents/pkg/agent/agent.go`):
```go
type Agent interface {
    ID() string
    Chat(ctx context.Context, prompt string, opts ...map[string]any) (*response.Response, error)
    ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *response.StreamingResponse, error)
    // ... prompt is always string
}
```

Changes:
- `prompt string` → `messages []protocol.Message` — aligns with raw LLM API contract, caller owns context management
- go-agents internally calls `initMessages(prompt)` to build `[system, user]` array. tau/agent does NOT do this — it passes messages directly to the format layer.

**Construction change**:

Kernel `New()`:
```go
func New(cfg *config.AgentConfig) (Agent, error)
// internally resolves provider from registry: providers.Create(cfg.Provider)
```

go-agents `New()`:
```go
func New(cfg *config.AgentConfig) (Agent, error)
// also internally resolves provider and format from registries
```

tau/agent `New()` — explicit injection:
```go
func New(cfg *config.AgentConfig, prov provider.Provider, fmt format.Format) (Agent, error)
// caller resolves from registries and passes in
```

**Error types**: Kernel and go-agents have identical `AgentError` types (same lineage). Port from either source.

### request/ Package

**Adapt from go-agents**: `~/code/go-agents/pkg/request/`

| go-agents Source | Destination | Lines | Action |
|------------------|-------------|-------|--------|
| `~/code/go-agents/pkg/request/interface.go` (32 lines) | `request/interface.go` | Request interface | Port with import changes |
| `~/code/go-agents/pkg/request/chat.go` (74 lines) | `request/chat.go` | ChatRequest | Port with import changes |
| `~/code/go-agents/pkg/request/vision.go` (83 lines) | `request/vision.go` | VisionRequest | Port with import changes |
| `~/code/go-agents/pkg/request/tools.go` (79 lines) | `request/tools.go` | ToolsRequest | Port with import changes |
| `~/code/go-agents/pkg/request/embeddings.go` (74 lines) | `request/embeddings.go` | EmbeddingsRequest | Port with import changes |

**Kernel source** (replaced): `~/tau/kernel/agent/request/` — 5 request files + interface + audio. Kernel requests call `provider.Marshal()` which no longer exists. go-agents requests call `format.Marshal()` which is the correct pattern.

**Deviations from go-agents source**:
- Import path changes only (`go-agents` → `tau/*` modules)
- go-agents Request interface has `Provider()`, `Format()`, `Model()` accessors — preserved

**Deviations from kernel source**:
- Kernel's `request.Request` interface has no `Format()` or `Model()` method — go-agents' richer interface adopted
- Kernel requests call `provider.Marshal(protocol, model, data)` — replaced by `format.Marshal(protocol, data)` (provider doesn't marshal)
- Kernel's `AudioRequest` not ported (deferred)

### client/ Package

**Merge kernel + go-agents**: Both have HTTP client with retry logic.

| Source | Destination | Lines | Action |
|--------|-------------|-------|--------|
| `~/code/go-agents/pkg/client/client.go` (282 lines) | `client/client.go` | Client interface + impl | Port as primary |
| `~/code/go-agents/pkg/client/retry.go` (149 lines) | `client/retry.go` | Retry logic | Port intact |
| `~/tau/kernel/agent/client/client.go` (260 lines) | reference | Compare for feature parity | |
| `~/tau/kernel/agent/client/retry.go` (149 lines) | reference | Compare for feature parity | |

**Deviations from go-agents source**:
- Import paths change to tau/* modules
- Client.Execute uses `format.Parse()` — same pattern, different import
- Client.ExecuteStream uses `provider.Stream().ReadStream()` then `format.ParseStreamChunk()` — same pattern

**Deviations from kernel source**:
- Kernel client returns `any` from Execute and channels `any` from ExecuteStream — replaced by typed `*response.Response` and `<-chan *response.StreamingResponse`
- Kernel client calls `provider.ProcessResponse()` / `provider.ProcessStreamResponse()` — replaced by `format.Parse()` / `format.ParseStreamChunk()`
- Kernel client has no streaming via StreamReader — go-agents streaming integration is new

### mock/ Package

**Merge kernel + go-agents**: go-agents has additional `MockFormat` that kernel lacks.

| Source | Destination | Lines | Action |
|--------|-------------|-------|--------|
| `~/code/go-agents/pkg/mock/agent.go` (243 lines) | `mock/agent.go` | MockAgent | Port with interface changes |
| `~/code/go-agents/pkg/mock/client.go` (106 lines) | `mock/client.go` | MockClient | Port with import changes |
| `~/code/go-agents/pkg/mock/provider.go` (183 lines) | `mock/provider.go` | MockProvider | Port with import changes |
| `~/code/go-agents/pkg/mock/format.go` (110 lines) | `mock/format.go` | MockFormat | Port (new to TAU) |
| `~/code/go-agents/pkg/mock/helpers.go` (107 lines) | `mock/helpers.go` | Convenience constructors | Port with adaptations |

**Deviations from go-agents**:
- MockAgent must implement `[]Message` interface instead of `string` prompt
- Import paths change

**Deviations from kernel**:
- Kernel mock (`~/tau/kernel/agent/mock/`) has no MockFormat — added from go-agents
- Kernel MockAgent returns `*response.ChatResponse` etc. — updated to `*response.Response`
- Kernel helpers (`NewSimpleChatAgent` etc.) return kernel response types — updated

### registry/ Package

**Port from kernel**: `~/tau/kernel/agent/registry.go` (161 lines)

| Kernel Source | Destination | Action |
|---------------|-------------|--------|
| `~/tau/kernel/agent/registry.go` | `registry/registry.go` | Port with adaptations |

**go-agents has no registry** — this is a kernel innovation preserved in tau/agent.

**Deviations from kernel**:
- Moves from root agent package to `registry/` sub-package
- `Registry` type contains a reference to format and provider registries (or uses their package-level Create functions) to construct agents during lazy instantiation
- `Register(name, cfg)` stores config; `Get(name)` calls `format.Create()`, `provider.Create()`, then `agent.New(cfg, prov, fmt)`
- Import paths change

## Dependencies

- `github.com/tailored-agentic-units/protocol` — Protocol constants, Message, Response, config, streaming
- `github.com/tailored-agentic-units/format` — Format interface, data types, ToolDefinition
- `github.com/tailored-agentic-units/provider` — Provider interface, transport
- `github.com/google/uuid` — Agent ID (UUIDv7)

## Integration Points

- **tau/kernel** creates agents and uses them in the runtime loop
- **tau/orchestrate** hub accepts agents as Participants (agent.Agent satisfies hub.Participant via ID())
