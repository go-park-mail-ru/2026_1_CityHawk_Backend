#!/usr/bin/env python3
import argparse
import json
from urllib.parse import urlencode


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://localhost:8080")
    parser.add_argument("--count", type=int, default=20_000)
    parser.add_argument("--cookie-header", default="")
    args = parser.parse_args()

    queries = [
        {"limit": "12", "offset": "0", "sort": "dateAsc"},
        {"limit": "12", "offset": "120", "sort": "dateAsc"},
        {"limit": "12", "offset": "1200", "sort": "dateAsc"},
        {"limit": "12", "offset": "0", "sort": "popular"},
        {"query": "Perf Event 09", "limit": "12", "offset": "0"},
        {"cityId": "10000000-0000-0000-0000-000000000001", "limit": "12", "offset": "0"},
        {"categoryId": "20000000-0000-0000-0000-000000000001", "limit": "12", "offset": "0"},
        {"tag": "30000000-0000-0000-0000-000000000001", "limit": "12", "offset": "0"},
    ]

    for i in range(args.count):
        query = queries[i % len(queries)]
        target = {
            "method": "GET",
            "url": f"{args.base_url}/api/events?{urlencode(query)}",
            "header": {},
        }
        if args.cookie_header:
            target["header"]["Cookie"] = [args.cookie_header]
        print(json.dumps(target, separators=(",", ":")))


if __name__ == "__main__":
    main()
