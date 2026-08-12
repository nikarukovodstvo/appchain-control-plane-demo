# Appchain Control Plane Demo

A small, production-shaped Go + React vertical slice for an EVM appchain deployment control plane. The data is synthetic and the project does not deploy a chain, touch a wallet, or handle production credentials.

## What it proves

- strict request validation and unknown-field rejection;
- required idempotency keys with deterministic IDs;
- safe replay and conflict behavior;
- concurrent in-memory state protected by a mutex;
- bounded HTTP bodies, server timeouts, security headers, and explicit JSON errors;
- a typed React admin slice showing success, replay, and failure state.

## Run the API

```bash
go test ./...
go run ./cmd/server
```

`go test -race ./...` is also supported on systems with a C toolchain available.

Create a configuration:

```bash
curl -i -X POST http://localhost:8080/v1/chain-configs \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: demo-001" \
  -d '{"name":"orders-l2","chainId":90101,"rpcUrl":"https://rpc.example.test","blockTimeMs":2000,"sequencerMode":"committee"}'
```

## Run the web slice

```bash
cd web
npm install
npm run build
npm run dev
```

During local development, configure the Vite proxy or serve the built assets behind the Go API. Production evolution would add PostgreSQL transactions, authentication/RBAC, an outbox for asynchronous provisioning, OpenTelemetry, migrations, and deployment adapters behind a least-privilege boundary.

## Acceptance checklist

- First valid request returns `201` and a configuration ID.
- Exact retry returns `200`, the same ID, and `Idempotency-Replayed: true`.
- Same key with changed input returns `409`.
- Invalid URL, chain ID, block time, mode, or unknown JSON field is rejected.
- Concurrent identical requests produce one logical record.

Built as a transparent technical work sample by Mikhail Nikitenkov. No client work or confidential source code is represented.
