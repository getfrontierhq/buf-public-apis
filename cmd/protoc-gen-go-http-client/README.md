# protoc-gen-go-http-client

A protoc plugin that generates HTTP client code with interfaces from proto service definitions with `google.api.http` annotations.

## Features

- Generates Go HTTP client code from proto services
- Automatic interface generation for all services and clients
- Type-safe HTTP method handling (GET, POST, PUT, DELETE, PATCH)
- Path parameter handling with automatic field mapping
- Private fields with public getter methods for better encapsulation
- Compile-time interface implementation checks
- Support for nested service structures

## Installation

```bash
go install github.com/getfrontierhq/buf-public-apis/cmd/protoc-gen-go-http-client@latest
```

## Usage

### 1. Add proto dependencies

Add to your `buf.yaml`:

```yaml
deps:
  - buf.build/googleapis/googleapis
```

### 2. Configure the plugin

Add to your `buf.gen.yaml`:

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

**Configuration Parameters:**
- `client=<proto_package>:<output_subdir>` - Required. Specifies the proto package and output directory
  - Example: `client=vendors.iniciador:vendors/iniciador/httpclient/client`
- `go_module_path=<path>` - Optional. Go module path for imports
  - Default: `github.com/getfrontierhq/schema/pkg/go`

### 3. Generate code

```bash
buf generate
```

## Generated Code Structure

### Interfaces (Base Names)

The plugin generates interfaces using the base service name:

```go
// Root client interface
type IniciadorClient interface {
    GetAccounts() AccountsService
    GetInvestments() InvestmentsClient
}

// Service interface
type AccountsService interface {
    GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error)
    ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error)
}
```

### Implementations (Impl Suffix)

Concrete structs implement the interfaces with "Impl" suffix:

```go
// Root client implementation
type IniciadorClientImpl struct {
    httpClient  *httpclient.HTTPClient
    accounts    *AccountsServiceImpl    // Private field
    investments *InvestmentsClientImpl  // Private field
}

// Getter methods return interface types
func (c *IniciadorClientImpl) GetAccounts() AccountsService {
    return c.accounts
}

// Constructor returns concrete type
func NewIniciadorClient(baseURL, token string) *IniciadorClientImpl {
    // ... initialization
}
```

### Compile-Time Checks

Generated code includes compile-time interface checks:

```go
var _ AccountsService = (*AccountsServiceImpl)(nil)
var _ IniciadorClient = (*IniciadorClientImpl)(nil)
```

## Usage Example

```go
package main

import (
    "context"
    "github.com/your-org/schema/pkg/go/vendors/iniciador/httpclient/client"
)

func main() {
    // Create client (returns *IniciadorClientImpl)
    c := client.NewIniciadorClient("https://api.example.com", "your-token")

    // Use services via getter methods (returns interface types)
    accounts := c.GetAccounts()  // Returns AccountsService interface
    resp, err := accounts.ListAccounts(context.Background(), &pb.ListAccountsRequest{
        Id: "link-id",
    })
    if err != nil {
        panic(err)
    }

    // Use nested services
    investments := c.GetInvestments()  // Returns InvestmentsClient interface
    titles := investments.GetTreasureTitlesService()  // Returns TreasureTitlesService interface
    titlesResp, err := titles.ListInvestments(context.Background(), &pb.ListInvestmentsRequest{
        LinkId: "link-id",
    })
}
```

## Testing with Interfaces

The generated interfaces make mocking easy:

```go
type MockAccountsService struct {
    ListAccountsFunc func(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error)
}

func (m *MockAccountsService) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
    if m.ListAccountsFunc != nil {
        return m.ListAccountsFunc(ctx, req)
    }
    return &pb.ListAccountsResponse{}, nil
}

// Implement other methods...

func TestMyFunction(t *testing.T) {
    mock := &MockAccountsService{
        ListAccountsFunc: func(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
            return &pb.ListAccountsResponse{
                Data: []*pb.Account{{AccountId: "test"}},
            }, nil
        },
    }

    // Use mock where AccountsService interface is expected
    var service client.AccountsService = mock
    // ... test code
}
```

## Proto Annotations

Services must have `google.api.http` annotations:

```protobuf
syntax = "proto3";

import "google/api/annotations.proto";

service AccountsService {
  rpc GetAccount(GetAccountRequest) returns (GetAccountResponse) {
    option (google.api.http) = {
      get: "/v1/data/links/{id}/data/accounts/{account_id}"
    };
  }

  rpc CreateAccount(CreateAccountRequest) returns (CreateAccountResponse) {
    option (google.api.http) = {
      post: "/v1/data/accounts"
      body: "*"
    };
  }
}
```

## Design Decisions

- **Interface naming**: Base service name (e.g., `AccountsService`)
- **Implementation naming**: Base name + "Impl" suffix (e.g., `AccountsServiceImpl`)
- **Constructor returns concrete type**: `NewIniciadorClient()` returns `*IniciadorClientImpl`
- **Getters return interfaces**: All getter methods return interface types for testability
- **Private fields**: Service fields are private, accessed only via getters
- **Getter naming**: All getters use `Get<Name>` prefix (e.g., `GetAccounts()`)

## Requirements

- Go 1.21+
- protoc-gen-star v2
- google.golang.org/protobuf

## License

Apache 2.0
