# HW 4 Performance Test Summary

Main entity: `event`.

Write endpoint:

```text
POST /api/events
```

Read endpoint:

```text
GET /api/events
```

## Data Volume

Final DB check:

```sql
SELECT count(*) FROM event WHERE title LIKE 'Perf Event %';
```

```text
count = 100000
```

## Load Test Results

| Scenario | Requests | Rate | Throughput | p50 | p95 | p99 | Success | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| baseline-create | 100000 | 200.00 | 181.03 | 1.84s | 4.62s | 5.20s | 91.13% | Write path overloaded at RATE=200 |
| baseline-create-fill | 2 | 2.00 | 1.98 | 7.68ms | 7.78ms | 7.78ms | 100.00% | Filled missing events |
| baseline-read | 5000 | 20.00 | 0.00 | 30.00s | 30.00s | 30.00s | 0.00% | Timed out before optimization |
| optimized-read | 5000 | 20.00 | 0.00 | 30.00s | 30.00s | 30.00s | 0.00% | Index-only optimization insufficient |
| optimized-read-v2 | 5000 | 20.00 | 0.01 | 30.00s | 30.00s | 30.00s | 0.04% | Pagination-first query rewrite still saturated |

## DB Stats Snapshot

| Metric | Baseline | Optimized indexes |
| --- | ---: | ---: |
| events_count | 100106 | 100106 |
| event_sessions_count | 100100 | 100100 |
| event_images_count | 100211 | 100211 |
| top query total_exec_time | 33233.08ms | 29920.14ms |
| top query mean_exec_time | 6.65ms | 5.98ms |
| deadlocks | 0 | 0 |

## Conclusion

The write workload created the required 100k main entities. The read endpoint
became the bottleneck: with 100k events, `GET /api/events` timed out at the
selected read load. Adding indexes helped PostgreSQL use more targeted access
paths, but did not make the endpoint stable. A second optimization changed the
query to select the page of event IDs first and load related data only for that
page, but the service still saturated on the VM. Further work should focus on
removing expensive count/sort work from the hot read path and introducing a
precomputed read model or cache for event cards.

