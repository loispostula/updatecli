"""Reads, or waits for, the GitHub GraphQL budget of the credentials updatecli runs with.

usage: ratelimit.py show
       ratelimit.py wait MIN_REMAINING

Uses UPDATECLI_GITHUB_TOKEN, or else the UPDATECLI_GITHUB_APP_* variables, like updatecli itself.
"""

import json
import os
import sys
import time
import urllib.request
from datetime import UTC, datetime
from pathlib import Path

import jwt


def _request(url: str, token: str, body: dict | None = None) -> dict:
    request = urllib.request.Request(
        url,
        method="POST",
        data=json.dumps(body).encode() if body is not None else None,
        headers={"Authorization": f"Bearer {token}", "Accept": "application/vnd.github+json"},
    )
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)


def _token() -> str:
    if token := os.environ.get("UPDATECLI_GITHUB_TOKEN"):
        return token
    key = os.environ.get("UPDATECLI_GITHUB_APP_PRIVATE_KEY") or Path(
        os.environ["UPDATECLI_GITHUB_APP_PRIVATE_KEY_PATH"]
    ).read_text()
    now = int(time.time())
    app_jwt = jwt.encode(
        {"iat": now - 60, "exp": now + 540, "iss": os.environ["UPDATECLI_GITHUB_APP_CLIENT_ID"]}, key, algorithm="RS256"
    )
    installation_id = os.environ["UPDATECLI_GITHUB_APP_INSTALLATION_ID"]
    return _request(f"https://api.github.com/app/installations/{installation_id}/access_tokens", app_jwt)["token"]


def rate_limit(token: str) -> dict:
    data = _request("https://api.github.com/graphql", token, {"query": "{ rateLimit { limit remaining used resetAt } }"})
    return data["data"]["rateLimit"] | {"at": datetime.now(UTC).isoformat()}


def main() -> None:
    token = _token()
    if sys.argv[1] == "show":
        print(json.dumps(rate_limit(token)))
        return

    minimum = int(sys.argv[2])
    while (current := rate_limit(token))["remaining"] < minimum:
        reset = datetime.fromisoformat(current["resetAt"].replace("Z", "+00:00"))
        wait = max(30, int((reset - datetime.now(UTC)).total_seconds()) + 15)
        print(f"GraphQL budget {current['remaining']}/{current['limit']}, waiting {wait}s for the reset", file=sys.stderr)
        time.sleep(wait)
    print(json.dumps(current))


if __name__ == "__main__":
    main()
