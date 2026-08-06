#!/usr/bin/env python3
"""Scan jam-conformance fuzz JSON for large keyvals (likely PVM program blobs).

Heuristic: hex-decoded value length > 20000 octets ≈ MetaCode preimage / program blob.
Prints candidate trace folders with unique blob digests (sha256 of value bytes).
"""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
from collections import defaultdict
from pathlib import Path


def hex_to_len(value: str) -> int:
    s = value[2:] if value.startswith(("0x", "0X")) else value
    return len(s) // 2


def hex_to_bytes(value: str) -> bytes:
    s = value[2:] if value.startswith(("0x", "0X")) else value
    return bytes.fromhex(s)


def scan_json(path: Path, min_octets: int) -> list[bytes]:
    try:
        data = json.loads(path.read_text())
    except Exception as e:
        print(f"skip {path}: {e}", file=sys.stderr)
        return []
    blobs: list[bytes] = []
    for side in ("pre_state", "post_state"):
        state = data.get(side) or {}
        for kv in state.get("keyvals") or []:
            val = kv.get("value")
            if not isinstance(val, str):
                continue
            if hex_to_len(val) <= min_octets:
                continue
            blobs.append(hex_to_bytes(val))
    return blobs


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument(
        "traces_root",
        nargs="?",
        default="pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces",
    )
    ap.add_argument("--min-octets", type=int, default=20000)
    ap.add_argument("--limit-folders", type=int, default=0, help="0 = all folders with blobs")
    ap.add_argument("--json-out", default="", help="optional JSON summary path")
    args = ap.parse_args()

    root = Path(args.traces_root)
    if not root.is_dir():
        print(f"ERROR: traces root not found: {root}", file=sys.stderr)
        return 1

    folder_blobs: dict[str, set[str]] = defaultdict(set)
    global_blobs: set[str] = set()

    for folder in sorted(p for p in root.iterdir() if p.is_dir()):
        for jf in sorted(folder.glob("*.json")):
            for blob in scan_json(jf, args.min_octets):
                digest = hashlib.sha256(blob).hexdigest()
                folder_blobs[folder.name].add(digest)
                global_blobs.add(digest)

    ranked = sorted(folder_blobs.items(), key=lambda kv: (-len(kv[1]), kv[0]))
    if args.limit_folders > 0:
        ranked = ranked[: args.limit_folders]

    print(f"unique_program_blobs={len(global_blobs)} min_octets={args.min_octets}")
    print(f"folders_with_blobs={len(folder_blobs)}")
    for name, digests in ranked:
        print(f"{name}\tunique_blobs={len(digests)}")

    if args.json_out:
        payload = {
            "unique_program_blobs": len(global_blobs),
            "folders": [
                {"folder": name, "unique_blobs": sorted(digests)}
                for name, digests in ranked
            ],
        }
        Path(args.json_out).write_text(json.dumps(payload, indent=2) + "\n")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
