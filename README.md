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

1.  **Install the Tool**:
    ```bash
    go install github.com/jinuthankachan/sdm/cmd/sdm@latest
    ```

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

Run the `sdm` tool:

```bash
sdm generate proto/invoice/invoice.proto
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

## Generated Schema Structure

*   **`pii_<name>s`**: Stores `pii` fields and `primary_key`.
*   **`chain_<name>s`**: key-value store for non-pii and `hashed` fields (EAV pattern).
*   **`<name>s` (View)**: Joins the PII table with the latest values from the Chain table.
