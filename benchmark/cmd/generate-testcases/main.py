#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = []
# ///
"""
HAR to test cases converter
Usage: uv run main.py traffic.har -o testdata/http_traffic.json
"""

import argparse
import json
import random
from typing import Dict, List
from urllib.parse import urlparse

RANDOM_SEED = 42
EXCLUDE_PATTERNS = [
    ".woff",
    ".woff2",
    ".ttf",
    ".eot",
    "favicon",
    "analytics",
    "tracking",
    "ads",
    "doubleclick",
    "googletagmanager",
]

CONTENT_TYPE_DISTRIBUTION = {
    "json": 0.30,
    "html": 0.10,
    "javascript": 0.25,
    "css": 0.10,
    "image": 0.15,
    "other": 0.10,
}


def categorize_content_type(mime_type: str) -> str:
    """Categorize MIME type for stratified sampling"""
    mime_lower = mime_type.lower()
    if "json" in mime_lower:
        return "json"
    if "html" in mime_lower:
        return "html"
    if "javascript" in mime_lower:
        return "javascript"
    if "css" in mime_lower:
        return "css"
    if "image" in mime_lower:
        return "image"
    return "other"


def should_exclude(url: str, patterns: List[str]) -> bool:
    """Check if URL should be excluded"""
    url_lower = url.lower()
    return any(pattern.lower() in url_lower for pattern in patterns)


def select_stratified(entries: List[Dict], limit: int, seed: int) -> List[Dict]:
    """Select entries using stratified sampling"""
    random.seed(seed)

    categories = {}
    for entry in entries:
        mime = entry["response"]["content"].get("mimeType", "")
        if ";" in mime:
            mime = mime.split(";")[0]

        category = categorize_content_type(mime)
        if category not in categories:
            categories[category] = []
        categories[category].append(entry)

    print("Content Type Distribution:")
    for cat, items in sorted(categories.items()):
        print(f"  {cat:20s} {len(items):4d} entries")
    print()

    targets = {cat: int(limit * pct) for cat, pct in CONTENT_TYPE_DISTRIBUTION.items()}

    total_targeted = sum(targets.values())
    if total_targeted < limit:
        targets["json"] += limit - total_targeted

    print("Target Distribution:")
    for cat, target in sorted(targets.items()):
        pct = CONTENT_TYPE_DISTRIBUTION.get(cat, 0) * 100
        print(f"  {cat:20s} {target:4d} ({pct:.0f}%)")
    print()

    selected = []
    for cat, target in targets.items():
        available = categories.get(cat, [])
        if not available:
            print(f"No entries for category '{cat}', skipping")
            continue

        take = min(target, len(available))
        if len(available) < target:
            print(f"Only {take} entries available for '{cat}' (wanted {target})")

        sampled = random.sample(available, take)
        selected.extend(sampled)

    # Fill deficit
    if len(selected) < limit:
        deficit = limit - len(selected)
        print(f"\nDeficit of {deficit} entries, filling randomly")
        all_entries = [e for items in categories.values() for e in items]
        remaining = [e for e in all_entries if e not in selected]
        additional = random.sample(remaining, min(deficit, len(remaining)))
        selected.extend(additional)

    return selected


def convert_entry(entry: Dict, index: int) -> Dict:
    """Convert HAR entry to test case"""
    url = entry["request"]["url"]
    parsed = urlparse(url)

    host = parsed.netloc
    path = parsed.path + ("?" + parsed.query if parsed.query else "") or "/"

    req_headers = {}
    for h in entry["request"]["headers"]:
        name = h["name"].lower()
        if name.startswith(":"):  # Skip HTTP/2 pseudo-headers
            continue
        req_headers[name] = h["value"]

    resp_headers = {}
    for h in entry["response"]["headers"]:
        name = h["name"].lower()
        if name.startswith(":"):
            continue
        resp_headers[name] = h["value"]

    method = entry["request"]["method"]
    status = entry["response"]["status"]

    return {
        "name": f"Request {index}: {method} {path}",
        "description": f"{method} {host}{path} - Status {status}",
        "request": {
            "method": method,
            "host": host,
            "path": path,
            "headers": req_headers,
        },
        "response": {
            "statusCode": status,
            "headers": resp_headers,
        },
    }


def main():
    parser = argparse.ArgumentParser(description="Extract HAR to test cases")
    parser.add_argument("har_file", help="Input HAR file")
    parser.add_argument(
        "-o", "--output", default="testdata/http_traffic.json", help="Output JSON file"
    )
    parser.add_argument(
        "-n", "--limit", type=int, default=100, help="Number of test cases to extract"
    )

    args = parser.parse_args()

    print(f"Loading HAR file: {args.har_file}")
    with open(args.har_file) as f:
        har = json.load(f)

    entries = har["log"]["entries"]
    print(f"Total entries: {len(entries)}")

    filtered = [
        e for e in entries if not should_exclude(e["request"]["url"], EXCLUDE_PATTERNS)
    ]
    print(f"After filtering (exclude patterns e.g. .ttf, .wof, ...): {len(filtered)}\n")

    selected = select_stratified(filtered, args.limit, RANDOM_SEED)
    print(f"\nSelected {len(selected)} entries")

    test_cases = []
    for i, entry in enumerate(selected, 1):
        try:
            tc = convert_entry(entry, i)
            test_cases.append(tc)
        except Exception as e:
            print(f"Warning: Failed to convert entry {i}: {e}")

    with open(args.output, "w") as f:
        json.dump(test_cases, f, indent=2)

    print(f"\nSuccessfully wrote {len(test_cases)} test cases to {args.output}")


if __name__ == "__main__":
    main()
