# Buf Plugins

A collection of protobuf plugins and annotations for Frontier Technologies.

## Plugins

### protoc-gen-go-dynamo

A protoc plugin that adds DynamoDB struct tags to generated Go protobuf code.

#### Installation

```bash
go install github.com/getfrontierhq/buf-public-apis/cmd/protoc-gen-go-dynamo@latest
```

#### Usage

1. Add the proto dependency to your `buf.yaml`:

```yaml
deps:
  - buf.build/frontier/public-apis
```

2. Import and use the annotations in your proto files:

```protobuf
syntax = "proto3";

import "dynamo/annotations.proto";

message User {
  string id = 1 [(dynamo.key) = {type: KEY_TYPE_HASH, column_name: "ID"}];
  string email = 2 [
    (dynamo.key) = {column_name: "email"},
    (dynamo.gsi) = {name: "email-index", key: KEY_TYPE_HASH}
  ];
  int64 created_at = 3 [(dynamo.key) = {type: KEY_TYPE_RANGE}];
}
```

3. Configure the plugin in your `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - local: protoc-gen-go-dynamo
    out: gen/go
    opt:
      - paths=source_relative
      - outdir=gen/go
```

4. Generate:

```bash
buf generate
```

#### Annotations

##### (dynamo.key)

Primary table key annotation using `KeyConfig` message.

- `type`: `KEY_TYPE_HASH` (partition key) or `KEY_TYPE_RANGE` (sort key)
- `column_name`: DynamoDB column name (optional)

Examples:
- `[(dynamo.key) = {type: KEY_TYPE_HASH, column_name: "ID"}]` → `` `dynamo:"ID,hash"` ``
- `[(dynamo.key) = {type: KEY_TYPE_RANGE}]` → `` `dynamo:",range"` ``
- `[(dynamo.key) = {column_name: "email"}]` → `` `dynamo:"email"` ``

##### (dynamo.gsi)

Global Secondary Index annotation using `IndexConfig` message (repeatable).

- `name`: Index name (required)
- `key`: `KEY_TYPE_HASH` or `KEY_TYPE_RANGE`

Examples:
- `[(dynamo.gsi) = {name: "email-index", key: KEY_TYPE_HASH}]` → `` `index:"email-index,hash"` ``

Multiple GSIs on one field:
```protobuf
string email = 1 [
  (dynamo.gsi) = {name: "email-index", key: KEY_TYPE_HASH},
  (dynamo.gsi) = {name: "secondary-index", key: KEY_TYPE_RANGE}
];
```

##### (dynamo.lsi)

Local Secondary Index annotation using `IndexConfig` message (repeatable).

Same format as GSI.

Example:
- `[(dynamo.lsi) = {name: "timestamp-index", key: KEY_TYPE_RANGE}]` → `` `localIndex:"timestamp-index,range"` ``

---

### protoc-gen-go-http

A protoc plugin that generates HTTP client and server handlers from `google.api.http` annotations.

#### Installation

```bash
go install github.com/getfrontierhq/buf-public-apis/cmd/protoc-gen-go-http@latest
```

#### Usage

1. Add the proto dependency to your `buf.yaml`:

```yaml
deps:
  - buf.build/googleapis/googleapis
```

2. Import and use annotations in your proto files:

```protobuf
syntax = "proto3";

import "google/api/annotations.proto";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/v1/users/{user_id}"
    };
  }

  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/v1/users"
      body: "*"
    };
  }
}
```

3. Configure the plugin in your `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - local: protoc-gen-go-http
    out: gen/go
    opt:
      - paths=source_relative
```

4. Generate:

```bash
buf generate
```

The plugin generates:
- HTTP handler registration functions
- Route binding code
- Request/response encoding/decoding

#### Runtime Library

The generated code requires the runtime library:

```bash
go get github.com/getfrontierhq/buf-public-apis/internal/gohttp
```

Use in your server:

```go
import (
  "github.com/go-chi/chi/v5"
  pb "your/generated/proto"
)

r := chi.NewRouter()
pb.RegisterUserServiceHTTPServer(r, yourService)
```

---

### protoc-gen-go-http-client

A protoc plugin that generates HTTP **client** code with automatic interface generation from `google.api.http` annotations.

#### Installation

```bash
go install github.com/getfrontierhq/buf-public-apis/cmd/protoc-gen-go-http-client@latest
```

#### Features

- HTTP client generation from proto services
- Automatic interface generation for testability
- Private fields with public getter methods
- Support for nested service structures
- Path parameter handling
- Compile-time interface checks

#### Usage

1. Add proto dependency:

```yaml
deps:
  - buf.build/googleapis/googleapis
```

2. Configure the plugin:

```yaml
version: v2
plugins:
  - local: protoc-gen-go-http-client
    out: pkg/go
    strategy: all
    opt:
      - paths=source_relative
      - client=vendors.iniciador:vendors/iniciador/httpclient/client
```

3. Generate:

```bash
buf generate
```

#### Example

```go
// Create client (returns *IniciadorClientImpl)
c := client.NewIniciadorClient("https://api.example.com", "token")

// Use services via getter methods (returns interface types)
accounts := c.GetAccounts()  // Returns AccountsService interface
resp, err := accounts.ListAccounts(ctx, req)
```

See [cmd/protoc-gen-go-http-client/README.md](cmd/protoc-gen-go-http-client/README.md) for detailed documentation.

---

## Proto Definitions

All proto annotations are published to:
- **BSR**: `buf.build/frontier/public-apis`

## License

Apache 2.0
<!-- attempt-version: 119 -->
