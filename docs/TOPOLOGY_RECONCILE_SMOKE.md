# Topology Reconcile Smoke Procedure

Manual verification for [#967](https://github.com/New-JAMneration/JAM-Protocol/issues/967) after validator node networking bootstrap (#964) and topology manager (#1029).

## Prerequisites

**Option A — automated harness (recommended for §1)**

```bash
./scripts/topology-smoke/run.sh
```

Uses temp keystores under `/tmp/jam-topology-smoke/` and `JAM_TOPOLOGY_SMOKE_CONFIG` to inject validator QUIC metadata (no chainspec editing).

**Option B — manual / testnet**

- Three or more validator Ed25519 keys in `keystore/ed25519/keys.json`
- A chainspec with validator metadata addresses (`bootnodes` or per-validator IP/port in metadata)
- Built node binary: `go build -o bin/node ./cmd/node`

## Local multi-validator layout

Run one validator node per key on distinct UDP ports, e.g.:

| Node | Role | Listen | Notes |
|------|------|--------|-------|
| A | validator | `127.0.0.1:10001` | index 0 |
| B | validator | `127.0.0.1:10002` | index 1 |
| C | validator | `127.0.0.1:10003` | index 2 |

Example start command:

```bash
./bin/node --role validator --listen-addr 127.0.0.1:10001 --chain <chainspec.json>
```

Ensure each node's metadata advertises the correct QUIC address so peers can dial.

## Automated 3-validator harness (FG-1)

For local §1 verification without polkajam testnet or editing `dev.chainspec.json` genesis state:

```bash
./scripts/topology-smoke/setup.sh   # only: generate /tmp/jam-topology-smoke/{node0,1,2,smoke_config.json}
./scripts/topology-smoke/run.sh       # build node, start 3 validators, check logs (~20s)
```

| Component | Role |
|-----------|------|
| `scripts/topology-smoke/setup.sh` | Dev-chain trivial_seed(0..2) Ed25519 keys → per-node `keystore/ed25519/keys.json` |
| `smoke_config.json` | Maps each validator pubkey → `127.0.0.1:10101/10102/10103` |
| `JAM_TOPOLOGY_SMOKE_CONFIG` | Read at bootstrap; seeds kappa/lambda/gamma_k with QUIC metadata (`topology_smoke_config.go`) |
| `validator/metadata.go` | `EncodeUDPMetadata` / `ValidatorWithUDPAddr` helpers |

**Pass criteria (§1):** each node log shows `topology smoke: seeded 3 validators`, `safrole: connectivity applied`, and `TLS connection established` / `on peer updated: peer=127.0.0.1:1010*`.

Harness is **dev-only**; unset `JAM_TOPOLOGY_SMOKE_CONFIG` in production.

## What to observe

### 1. Transport connectivity reconcile

After all nodes start:

- Logs should show `StartValidatorConnections` dial attempts toward required peers (prev/current/next epoch validator set minus self).
- Preferred Initiator peers connect immediately; others may wait ~5s before dialing.
- `topology: block imported at slot ...` appears when blocks are imported via sync.

**Pass:** every pair in the required transport set has an established QUIC connection (no perpetual dial errors for reachable peers).

### 2. Stale peer prune

Stop one validator node while others keep running. Within ~15s (reconcile interval):

- Remaining nodes should log `topology: disconnect stale peer ...` for the stopped validator (builder connections are skipped).

**Pass:** stale validator connections are removed; builder-role connections are not pruned by topology.

### 3. Epoch transition delay

When the chain advances into a new epoch:

1. Topology records a pending epoch transition.
2. Before delay elapses, reconcile only **dials missing** peers; it does **not** refresh validator sets or re-open UP 0 gossip streams.
3. After both conditions hold:
   - new epoch's first block is finalized
   - `max(floor(E/30), 1)` slots since epoch start
4. Log: `topology: applying epoch N connectivity after delay`
5. `ConnectivityApplied` event is published (Safrole scheduler and UP 0 timing consume this).

**Pass:** connectivity / gossip changes apply only after the delay, not at the first block of the new epoch.

### 4. UP 0 grid gossip

On a validator with grid neighbors configured:

- UP 0 streams open only to grid-gossip-eligible peers (`ReconcileUP0Streams`).
- Non-neighbor validator connections remain for transport but do not keep UP 0 streams.

**Pass:** UP 0 stream count matches grid gossip set size, not full transport set.

### 5. Safrole timing (FG follow-up)

After `ConnectivityApplied`:

- CE 131 step-1 window opens after `max(floor(E/60), 1)` finalized slots.
- CE 132 forwarding waits until `max(floor(E/20), 1)` finalized slots (see `safrole` scheduler logs).

**Pass:** Safrole logs show step-1 window only after connectivity anchor + delay.

## Automated regression

```bash
go test ./internal/networking/topology/... -count=1
go test ./internal/networking/epochclock/... -count=1
go test ./internal/networking/safrole/... -count=1
go test ./internal/networking/validator/... -count=1
go test ./internal/networking/quic/... -count=1
```

Also run before merge:

```bash
go test ./internal/node/... -count=1
go test ./cmd/node/... -count=1
go build -o bin/node ./cmd/node
./scripts/topology-smoke/run.sh
```

## Smoke run log (2026-07-06)

| Step | Status | Notes |
|------|--------|-------|
| Automated regression (above) | PASS | All packages exit 0 |
| `go build -o bin/node ./cmd/node` | PASS | |
| Full node startup (`--role full`) | PASS | `node networking started`, graceful shutdown on SIGTERM |
| Validator startup (`--role validator` + temp keystore) | PASS | `safrole: connectivity applied for epoch 0`, topology manager starts |
| Multi-validator transport reconcile (§1) | PASS | `./scripts/topology-smoke/run.sh` — 3 Go validators, mock metadata, TLS to 10101/10102/10103 |
| Stale peer prune (§2) | NOT RUN | Needs running multi-validator cluster |
| Epoch transition delay (§3) | NOT RUN | Needs chain advancing across epoch boundary |
| UP 0 grid gossip (§4) | NOT RUN | Needs connected grid neighbours |
| Safrole step-1 window log (§5) | PARTIAL | `ConnectivityApplied` log seen at validator boot; step-1 window needs `BlockImported` advancing finalized slots |

## Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| No dials | Empty validator metadata address; check `PeerAddressFromMetadata` |
| Duplicate dials | Both sides dial; expected — QUIC allows single connection per pair |
| Epoch never applies | Finalized head not advancing; sync path (#567) not importing blocks |
| UP 0 never opens | Peer not in grid gossip set; check `GossipPeerKeys` / validator index |
