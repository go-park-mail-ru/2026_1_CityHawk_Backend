#!/usr/bin/env python3
import argparse
import base64
import json
from datetime import datetime, timedelta, timezone


CATEGORY_IDS = [
    "20000000-0000-0000-0000-000000000001",
    "20000000-0000-0000-0000-000000000002",
    "20000000-0000-0000-0000-000000000003",
    "20000000-0000-0000-0000-000000000004",
]

TAG_IDS = [
    "30000000-0000-0000-0000-000000000001",
    "30000000-0000-0000-0000-000000000002",
    "30000000-0000-0000-0000-000000000003",
    "30000000-0000-0000-0000-000000000004",
]


def place_id(index: int) -> str:
    return f"40000000-0000-0000-0000-{index:012d}"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://localhost:8080")
    parser.add_argument("--count", type=int, default=100_000)
    parser.add_argument("--start-index", type=int, default=0)
    parser.add_argument("--places", type=int, default=2048)
    parser.add_argument("--csrf-token", required=True)
    parser.add_argument("--cookie-header", required=True)
    args = parser.parse_args()

    start = datetime(2027, 1, 1, 9, 0, tzinfo=timezone.utc)

    for i in range(args.start_index, args.start_index + args.count):
        current_place = (i % args.places) + 1
        slot = i // args.places
        starts_at = start + timedelta(hours=3 * slot)
        ends_at = starts_at + timedelta(hours=2)
        category = CATEGORY_IDS[i % len(CATEGORY_IDS)]
        tag = TAG_IDS[i % len(TAG_IDS)]

        body = {
            "title": f"Perf Event {i:06d}",
            "shortDescription": f"Generated event #{i} for DB performance testing",
            "fullDescription": (
                "Synthetic payload for HW 4. "
                "The text has stable length and realistic characters."
            ),
            "ageLimit": i % 22,
            "sourceUrl": "https://example.com/perf",
            "categoryIds": [category],
            "tagIds": [tag],
            "imageUrls": [f"https://example.com/perf/{i:06d}.jpg"],
            "placeId": place_id(current_place),
            "sessions": [
                {
                    "placeId": place_id(current_place),
                    "startAt": starts_at.isoformat().replace("+00:00", "Z"),
                    "endAt": ends_at.isoformat().replace("+00:00", "Z"),
                    "price": (i % 25) * 100,
                }
            ],
        }
        encoded = base64.b64encode(json.dumps(body, separators=(",", ":")).encode()).decode()
        target = {
            "method": "POST",
            "url": f"{args.base_url}/api/events",
            "header": {
                "Content-Type": ["application/json"],
                "X-CSRF-Token": [args.csrf_token],
                "Cookie": [args.cookie_header],
            },
            "body": encoded,
        }
        print(json.dumps(target, separators=(",", ":")))


if __name__ == "__main__":
    main()
