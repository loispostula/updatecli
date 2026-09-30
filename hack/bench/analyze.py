"""Compares the benchmark runs under runs/: wall time, clean pass, and GitHub GraphQL calls."""

import json
import re
from datetime import datetime
from pathlib import Path

RUNS = Path(__file__).parent / "runs"


def _log_times(log: Path) -> dict:
    marks, summary, counts = {}, {}, {"on_hold": 0, "rate_limited": 0, "errors": 0, "not_closed": 0}
    for line in log.open(errors="replace"):
        stamp, _, message = line.partition(" ")
        if "BENCHMARK: clean pass start" in message:
            marks["clean_start"] = datetime.fromisoformat(stamp)
        elif "BENCHMARK: clean pass end" in message:
            marks["clean_end"] = datetime.fromisoformat(stamp)
        elif "on hold for" in message:
            counts["on_hold"] += 1
        elif "rate limit exceeded" in message.lower():
            counts["rate_limited"] += 1
        elif "BENCHMARK BUILD: not closing" in message:
            counts["not_closed"] += 1
        elif message.startswith("ERROR"):
            counts["errors"] += 1
        if m := re.search(r"\* (Changed|Failed|Skipped|Succeeded|Total):\s+(\d+)", message):
            summary[m.group(1)] = int(m.group(2))
    return marks | {"summary": summary} | counts


def _graphql_calls(spans: Path, clean_start: datetime | None, clean_end: datetime | None) -> dict:
    lo = clean_start.timestamp() * 1e9 if clean_start else None
    hi = clean_end.timestamp() * 1e9 if clean_end else None
    total = in_clean = 0
    busy_ns = 0
    for line in spans.open():
        for resource in json.loads(line)["resourceSpans"]:
            for scope in resource["scopeSpans"]:
                for span in scope["spans"]:
                    if span["name"] != "HTTP POST":
                        continue
                    url = next((a["value"].get("stringValue", "") for a in span.get("attributes", []) if a["key"] == "url.full"), "")
                    if "api.github.com/graphql" not in url:
                        continue
                    total += 1
                    start, end = int(span["startTimeUnixNano"]), int(span["endTimeUnixNano"])
                    if lo and hi and lo <= start <= hi:
                        in_clean += 1
                        busy_ns += end - start
    return {"graphql_total": total, "graphql_clean": in_clean, "graphql_clean_busy_s": round(busy_ns / 1e9, 1)}


def main() -> None:
    rows = []
    for run in sorted(RUNS.iterdir()):
        if not (run / "summary.json").exists():
            continue
        summary = json.loads((run / "summary.json").read_text())
        log = _log_times(run / "updatecli.log")
        calls = _graphql_calls(run / "spans.jsonl", log.get("clean_start"), log.get("clean_end"))
        clean_s = (log["clean_end"] - log["clean_start"]).total_seconds() if "clean_end" in log else None
        before = json.loads((run / "ratelimit-before.json").read_text())
        after = json.loads((run / "ratelimit-after.json").read_text())
        rows.append({
            "run": run.name,
            "exit": summary["exit"],
            "pipelines": log["summary"].get("Total"),
            "changed": log["summary"].get("Changed"),
            "failed": log["summary"].get("Failed"),
            "wall_min": round(summary["wall_s"] / 60, 1),
            "clean_s": round(clean_s, 1) if clean_s is not None else None,
            **calls,
            "on_hold": log["on_hold"],
            "rate_limited": log["rate_limited"],
            "not_closed": log["not_closed"],
            "errors": log["errors"],
            "budget_before": before["remaining"],
            "budget_after": after["remaining"],
            "reset_crossed": before["resetAt"] != after["resetAt"],
        })

    columns = list(rows[0]) if rows else []
    print(" | ".join(columns))
    for row in rows:
        print(" | ".join(str(row[c]) for c in columns))
    (RUNS / "results.json").write_text(json.dumps(rows, indent=2))


if __name__ == "__main__":
    main()
