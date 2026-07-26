# Topology Reconcile Smoke Procedure

> Note: This document describes a reusable local/manual verification workflow.
> Point-in-time run evidence should live in PR comments or the PR body, not in
> this repository document.

Manual verification procedure for topology reconcile, UP 0 stream selection,
and Safrole timing after validator node networking bootstrap.

## Prerequisites

**Option A - automated local harness (recommended for transport reconcile)**

```bash
./scripts/test-topology-smoke-run.sh
```

Uses temp keystores under `/tmp/jam-topology-smoke/` and
`JAM_TOPOLOGY_SMOKE_CONFIG` to inject validator QUIC metadata without editing
the chainspec.

**Option B - manual / testnet**

- Three or more validator Ed25519 keys in `keystore/ed25519/keys.json`
- A chainspec with validator metadata addresses (`bootnodes` or per-validator
  IP/port in metadata)
- Built node binary: `go build -o bin/node ./cmd/node`

## Local Multi-Validator Layout

Run one validator node per key on distinct UDP ports, for example:

| Node | Role | Listen | Notes |
|------|------|--------|-------|
| A | validator | `127.0.0.1:10001` | index 0 |
| B | validator | `127.0.0.1:10002` | index 1 |
| C | validator | `127.0.0.1:10003` | index 2 |

Example start command:

```bash
./bin/node --role validator --listen-addr 127.0.0.1:10001 --chain <chainspec.json>
```

Ensure each node's metadata advertises the correct QUIC address so peers can
dial.

## Automated 3-Validator Harness

For local transport-reconcile verification without editing genesis state:

```bash
./scripts/test-topology-smoke-setup.sh
./scripts/test-topology-smoke-run.sh
```

| Component | Role |
|-----------|------|
| `scripts/test-topology-smoke-setup.sh` | Generate per-node keystores and `smoke_config.json` |
| `scripts/test-topology-smoke-run.sh` | Build node, start 3 validators, and check logs |
| `JAM_TOPOLOGY_SMOKE_CONFIG` | Seeds kappa/lambda/gamma_k with QUIC metadata at bootstrap |
| `validator/metadata.go` | `EncodeUDPMetadata` / `ValidatorWithUDPAddr` helpers |

Pass criteria for transport reconcile:

- Each node logs `topology smoke: seeded 3 validators`
- Each node logs `safrole: connectivity applied`
- Logs show `TLS connection established` or `on peer updated: peer=127.0.0.1:1010*`

Harness is dev-only; unset `JAM_TOPOLOGY_SMOKE_CONFIG` in production.

## What To Observe

### 1. Transport Connectivity Reconcile

After all nodes start:

- Logs show `StartValidatorConnections` dial attempts toward required peers
- Preferred initiator peers connect immediately; others may wait about 5s
- `topology: block imported at slot ...` appears when blocks are imported via sync

Pass condition: every pair in the required transport set has an established
QUIC connection and no perpetual dial errors for reachable peers.

### 2. Stale Peer Prune

Stop one validator node while others keep running. Within about 15s:

- Remaining nodes log `topology: disconnect stale peer ...` for the stopped validator

Pass condition: stale validator connections are removed; builder-role
connections are not pruned by topology.

### 3. Epoch Transition Delay

When the chain advances into a new epoch:

1. Topology records a pending epoch transition.
2. Before delay elapses, reconcile only dials missing peers; it does not
   refresh validator sets or re-open UP 0 gossip streams.
3. After both conditions hold:
   - the new epoch's first block is finalized
   - `max(floor(E/30), 1)` slots since epoch start have elapsed
4. Log: `topology: applying epoch N connectivity after delay`
5. `ConnectivityApplied` is published.

Pass condition: connectivity and gossip changes apply only after the delay, not
at the first block of the new epoch.

### 4. UP 0 Grid Gossip

On a validator with grid neighbors configured:

- UP 0 streams open only to grid-gossip-eligible peers
- Non-neighbor validator connections remain for transport but do not keep UP 0 streams

Pass condition: UP 0 stream count matches the grid gossip set size, not the
full transport set.

### 5. Safrole Timing

After `ConnectivityApplied`:

- CE 131 step-1 window opens after `max(floor(E/60), 1)` finalized slots
- CE 132 forwarding waits until `max(floor(E/20), 1)` finalized slots

Pass condition: Safrole logs show step-1 window only after the connectivity
anchor plus delay.

## Suggested Regression Commands

```bash
go test ./internal/networking/topology/... -count=1
go test ./internal/networking/epochclock/... -count=1
go test ./internal/networking/safrole/... -count=1
go test ./internal/networking/validator/... -count=1
go test ./internal/networking/quic/... -count=1
go test ./internal/node/... -count=1
go test ./cmd/node/... -count=1
go build -o bin/node ./cmd/node
./scripts/test-topology-smoke-run.sh
```

## Troubleshooting

| Symptom | Likely cause |
|---------|--------------|
| No dials | Empty validator metadata address; check `PeerAddressFromMetadata` |
| Duplicate dials | Both sides dial; expected because QUIC allows one connection per pair |
| Epoch never applies | Finalized head not advancing; sync path is not importing blocks |
| UP 0 never opens | Peer not in grid gossip set; check `GossipPeerKeys` / validator index |
