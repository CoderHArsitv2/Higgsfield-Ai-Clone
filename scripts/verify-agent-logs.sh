#!/usr/bin/env bash
# The agent logs and CAPTURE-TEST.md are part of the submission, and
# CAPTURE-TEST.md makes claims about specific files. A commit that removes one
# turns those claims into lies silently, which is how five of them went missing
# once already -- so this is checked rather than trusted.
set -euo pipefail
cd "$(dirname "$0")/.."

fail=0

count=$(find .agent-logs -name '*.md' | wc -l | tr -d ' ')
if [ "$count" -lt 1 ]; then
  echo "::error::.agent-logs contains no session logs"
  fail=1
else
  echo "session logs present: $count"
fi

# Every .agent-logs path CAPTURE-TEST.md points at must exist.
missing=0
while read -r path; do
  [ -z "$path" ] && continue
  if [ ! -f "$path" ]; then
    echo "::error::CAPTURE-TEST.md references a missing file: $path"
    missing=$((missing + 1))
  fi
done < <(grep -oE '\.agent-logs/[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9-]+_[0-9a-f-]+\.md' CAPTURE-TEST.md | sort -u)

if [ "$missing" -gt 0 ]; then
  fail=1
else
  echo "every file CAPTURE-TEST.md references exists"
fi

# The logs must never be ignored: they ship with the repo.
if git check-ignore -q .agent-logs 2>/dev/null; then
  echo "::error::.agent-logs is gitignored; it is part of the submission"
  fail=1
fi

exit "$fail"
