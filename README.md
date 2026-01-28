# Sensitive Data Management (SDM)

SDM is a toolset for Golang projects to manage sensitive data (PII) by separating it from public chain data using Protobuf annotations. It automatically generates Go models, SQL schemas, and Repository functions to handle the data flow.

## Features

*   **Proto Annotations**: Define `primary_key`, `pii`, `hashed`, etc., directly in your `.proto` files.
*   **Auto-Generated Go Models**: Creates GORM-compatible structs for PII tables, Chain tables, and combined Views.
*   **Auto-Generated SQL**: Generates `CREATE TABLE` and `CREATE VIEW` statements for PostgreSQL.
*   **Auto-Generated Repositories**: Generates type-safe `Save` and `Fetch` methods that handle:
    *   Splitting data into PII and Chain tables.
    *   Hashing fields marked as `hashed`.
    *   Reconstructing objects from the DB View.

## Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/jinuthankachan/sdm.git
    cd sdm
    ```

2.  **Build the Plugin**:
    ```bash
    go build -o bin/protoc-gen-sdm ./cmd/protoc-gen-sdm
    ```

3.  **Ensure `protoc` is installed**: You need the Protocol Buffers compiler.

## Usage

### 1. Define your Data Model

Create a `.proto` file and import `proto/sdm/annotations.proto`. Annotate your fields:

```protobuf
syntax = "proto3";
package invoice;

import "proto/sdm/annotations.proto";

option go_package = "github.com/jinuthankachan/sdm/proto/invoice";

message Invoice {
  string id = 1 [(sdm.primary_key) = true, (sdm.chain_identifier_key) = true];
  int64 invoice_number = 2 [(sdm.pii) = true];
  string seller_gst = 3 [(sdm.pii) = true, (sdm.hashed) = true];
  // ...
}
```

### 2. Generate Code

Run `protoc` with the `protoc-gen-sdm` plugin:

```bash
export PATH=$PATH:$(pwd)/bin
protoc --plugin=bin/protoc-gen-sdm \
       --sdm_out=. --sdm_opt=paths=source_relative \
       --go_out=. --go_opt=paths=source_relative \
       -I . -I proto \
       proto/invoice/invoice.proto
```

This will generate:
*   `invoice_sdm_model.go`: Structs.
*   `invoice_sdm_schema.sql`: SQL DDL.
*   `invoice_sdm_repo.go`: Repository implementation.

### 3. Use in Go

```go
import (
    "context"
    "gorm.io/gorm"
    "github.com/jinuthankachan/sdm/proto/invoice"
)

func main() {
    db, _ := gorm.Open(...) 
    repo := invoice.NewInvoiceRepo(db)

    // Save (Splits and Hashes automatically)
    err := repo.Save(ctx, &invoice.Invoice{
        Id: "inv_123",
        InvoiceNumber: 1001,
        SellerGst: "GST001",
    })

    // Fetch (reconstructs from View)
    view, err := repo.Fetch(ctx, "inv_123")
}
```

## Using as a Tool / SDK in External Projects

To use SDM in your own Go project:

1.  **Install the plugin**:
    ```bash
    go install github.com/jinuthankachan/sdm/cmd/protoc-gen-sdm@latest
    ```

2.  **Import annotations**:
    *   Vendor the `sdm` dependency to make `annotations.proto` available for `protoc`.
    *   Example `go.mod`:
        ```go
        require github.com/jinuthankachan/sdm v0.0.0-xxxx
        ```
    *   Run `go mod vendor`.

3.  **Define your Proto**:
    ```protobuf
    import "github.com/jinuthankachan/sdm/proto/sdm/annotations.proto";
    ```

4.  **Generate**:
    ```bash
    protoc --plugin=protoc-gen-sdm \
           --sdm_out=. --sdm_opt=paths=source_relative \
           --go_out=. --go_opt=paths=source_relative \
           -I . -I vendor/github.com/jinuthankachan/sdm \
           path/to/your.proto
    ```

Check the `example/demo` directory for a complete working example.

## Generated Schema Structure

*   **`pii_<name>s`**: Stores `pii` fields and `primary_key`.
*   **`chain_<name>s`**: key-value store for non-pii and `hashed` fields (EAV pattern).
*   **`<name>s` (View)**: Joins the PII table with the latest values from the Chain table.
