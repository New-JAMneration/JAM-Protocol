# GP v0.8.0 Phase Verification and Fixes

## Evidence audit

- [x] Read #1014, #1016, and #1017 issue text and attached equation-diff images.
- [x] Read PR #1027, #1031, and #1032 descriptions, reviews, comments, and follow-up commits.
- [x] Verify requirements against the official Gray Paper v0.8.0 TeX.
- [x] Separate current fixes from #1037 and official-vector follow-ups.

## #1016 Reporting and assurances

- [x] Validate `erasure_shards` against the posterior active validator-set length.
- [x] Validate anchor timeslot against the matching recent-history entry.
- [x] Validate lookup-anchor state root against the retained ancestor header.
- [x] Reject guarantee credentials outside the GP-required range of two to three.
- [x] Apply offender key replacement without rejecting every guarantee.
- [x] Add targeted validation and transition tests.

## #1014 Header and recent history

- [x] Populate ancestry when the cache is initially empty.
- [x] Keep the retained ancestry bounded and deterministic.
- [x] Add epoch/tickets marker presence and absence tests.
- [x] Add recent-history correction, append, timeslot, and eviction tests.
- [x] Add ancestry initialization and update tests.

## #1017 Disputes and judgments

- [x] Reject faults whose target has neither a good nor bad verdict.
- [x] Require each fault vote to contradict the target verdict.
- [x] Add good, bad, wonky, and unknown-target tests.
- [x] Add active- and previous-validator-set signature-path tests.
- [x] Verify offender-state updates in regression tests.

## Deferred scope

- [x] Track variable validator-set cardinality, thresholds, inactive cores, and codec sizing in #1037.
- [x] Track reflection registries, `jam_types` schema, official error enum, and full STF vectors with official v0.8.0 vectors.

## Verification and PR

- [x] Run targeted Docker tests for each phase.
- [x] Run related package tests with `-race`.
- [x] Run scoped `go vet`, `gofmt -s`, and Linux Docker build.
- [x] Commit each phase independently and update PR #1042.
