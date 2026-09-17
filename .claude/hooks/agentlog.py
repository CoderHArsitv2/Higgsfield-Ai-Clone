"""Shared helpers for the 8x assignment agent-capture hooks.

Writes one markdown file per Claude Code session into .agent-logs/, in the
format specified by the assignment brief. Captures ONLY the verbatim user
prompt and the final assistant text response for each turn -- no thinking,
no tool calls, no intermediate steps.
"""

import datetime
import fcntl
import json
import os
import re
import sys
import time

TOOL = "claude-code"
AUTHOR = "CoderHArsit"

# State lives outside the repo so it never pollutes the committed log dir.
STATE_DIR = os.path.expanduser("~/.claude/.agent-capture-state")

ENTRY_RE = re.compile(
    r"^\[LOG_ENTRY type=(?P<type>PROMPT|RESPONSE) num=(?P<num>\d+) session=\S+\]\n"
    r"timestamp: (?P<ts>\S+)\n"
    r"model: (?P<model>.*)$",
    re.MULTILINE,
)


def now_iso():
    return datetime.datetime.now(datetime.timezone.utc).strftime(
        "%Y-%m-%dT%H:%M:%S.") + f"{datetime.datetime.now(datetime.timezone.utc).microsecond // 1000:03d}Z"


def repo_root():
    # hooks live at <root>/.claude/hooks/
    return os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


def log_dir():
    d = os.path.join(repo_root(), ".agent-logs")
    os.makedirs(d, exist_ok=True)
    return d


def project_name():
    return os.path.basename(repo_root())


def debug_dump(event, payload):
    """Raw stdin dump, only when CAPTURE_DEBUG is set. Never on by default."""
    target = os.environ.get("CAPTURE_DEBUG")
    if not target:
        return
    try:
        os.makedirs(target, exist_ok=True)
        with open(os.path.join(target, f"{event}.json"), "a") as fh:
            fh.write(json.dumps(payload, indent=2) + "\n")
    except Exception:
        pass


def record_failure(event, raw, exc):
    """A capture hook that fails quietly is worse than no hook at all: the
    prompt is gone and nothing says so. Every failure leaves a record next to
    the logs, with the raw stdin, so a gap is always explainable."""
    try:
        path = os.path.join(log_dir(), "_capture-errors.log")
        with open(path, "a") as fh:
            fh.write(f"=== {now_iso()} {event} ===\n{exc}\nRAW STDIN:\n{raw}\n\n")
    except Exception:
        pass


def state_path(session_id):
    os.makedirs(STATE_DIR, exist_ok=True)
    return os.path.join(STATE_DIR, f"{session_id}.json")


def read_state(session_id):
    try:
        with open(state_path(session_id)) as fh:
            return json.load(fh)
    except Exception:
        return {}


def write_state(session_id, state):
    try:
        with open(state_path(session_id), "w") as fh:
            json.dump(state, fh)
    except Exception:
        pass


def session_file(session_id, create_ts=None):
    """Find the log file for this session, or create one named for its start."""
    d = log_dir()
    for name in sorted(os.listdir(d)):
        if name.endswith(f"_{session_id}.md"):
            return os.path.join(d, name)
    if create_ts is None:
        return None
    dt = datetime.datetime.strptime(create_ts[:19], "%Y-%m-%dT%H:%M:%S")
    fname = dt.strftime("%Y-%m-%d_%H-%M-%S") + f"_{session_id}.md"
    return os.path.join(d, fname)


def parse_entries(body):
    """Pull metadata out of already-written entries so the frontmatter can be
    rebuilt without ever touching entry text."""
    out = []
    for m in ENTRY_RE.finditer(body):
        out.append(
            {"type": m.group("type"), "num": int(m.group("num")),
             "ts": m.group("ts"), "model": m.group("model").strip()}
        )
    return out


def build_header(session_id, entries):
    prompts = [e for e in entries if e["type"] == "PROMPT"]
    models = []
    for e in entries:
        m = e["model"]
        if m and m not in models and m not in ("pending", "unknown"):
            models.append(m)
    first = prompts[0]["ts"] if prompts else (entries[0]["ts"] if entries else now_iso())
    last = prompts[-1]["ts"] if prompts else first
    date = first[:10]
    short = session_id.split("-")[0]
    fm = [
        "---",
        f"session_id: {session_id}",
        f"date: {date}",
        f"author: {AUTHOR}",
        f"model: {', '.join(models) if models else 'unknown'}",
        f"tool: {TOOL}",
        f"project: {project_name()}",
        f"total_exchanges: {len(prompts)}",
        f"first_prompt_time: {first}",
        f"last_prompt_time: {last}",
        "---",
        "",
        f"# Session Log - {date}",
        "",
        f"Session: `{short}` | Project: `{project_name()}` | Author: `{AUTHOR}`",
        "",
        "---",
        "",
        "",
    ]
    return "\n".join(fm)


def append_entry(session_id, kind, num, timestamp, model, text):
    """Append one LOG_ENTRY and regenerate the frontmatter. Existing entry text
    is never rewritten -- only the header block above the first entry."""
    path = session_file(session_id, create_ts=timestamp)
    short = session_id.split("-")[0]
    block = (
        f"[LOG_ENTRY type={kind} num={num} session={short}]\n"
        f"timestamp: {timestamp}\n"
        f"model: {model}\n\n"
        f"{text}\n\n\n"
    )

    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "a+") as fh:
        fcntl.flock(fh, fcntl.LOCK_EX)
        fh.seek(0)
        existing = fh.read()
        idx = existing.find("[LOG_ENTRY ")
        body = existing[idx:] if idx != -1 else ""
        body = body + block
        entries = parse_entries(body)
        fh.seek(0)
        fh.truncate()
        fh.write(build_header(session_id, entries) + body)
        fcntl.flock(fh, fcntl.LOCK_UN)
    return path


def next_prompt_num(session_id):
    path = session_file(session_id)
    if not path or not os.path.exists(path):
        return 1
    with open(path) as fh:
        entries = parse_entries(fh.read())
    return sum(1 for e in entries if e["type"] == "PROMPT") + 1


def current_prompt_num(session_id):
    path = session_file(session_id)
    if not path or not os.path.exists(path):
        return 1
    with open(path) as fh:
        entries = parse_entries(fh.read())
    prompts = [e for e in entries if e["type"] == "PROMPT"]
    return prompts[-1]["num"] if prompts else 1


def resolve_pending_model(session_id, num, model):
    """Fill in the model on a PROMPT entry that was written before the model was
    knowable (the first turn of a session, where no assistant message exists
    yet). This completes a field of the turn being written -- it never touches
    prompt or response text, and only ever replaces the literal 'pending'."""
    path = session_file(session_id)
    if not path or not os.path.exists(path) or not model or model == "pending":
        return
    with open(path, "r+") as fh:
        fcntl.flock(fh, fcntl.LOCK_EX)
        s = fh.read()
        marker = f"[LOG_ENTRY type=PROMPT num={num} session="
        i = s.find(marker)
        if i != -1:
            j = s.find("\n\n", i)
            head, rest = s[i:j], s[j:]
            if "model: pending" in head:
                s = s[:i] + head.replace("model: pending", f"model: {model}") + rest
                k = s.find("[LOG_ENTRY ")
                body = s[k:] if k != -1 else ""
                s = build_header(session_id, parse_entries(body)) + body
                fh.seek(0); fh.truncate(); fh.write(s)
        fcntl.flock(fh, fcntl.LOCK_UN)


def wait_for_model(transcript_path, timeout=4.0):
    """The Stop hook can fire fractionally before Claude Code flushes the final
    assistant message to the transcript, so the model is briefly unreadable.
    Poll rather than guess."""
    deadline = time.time() + timeout
    while True:
        rows = load_transcript(transcript_path)
        for row in reversed(rows):
            if row.get("type") == "assistant" and not row.get("isSidechain"):
                m = (row.get("message") or {}).get("model")
                if m:
                    return m, rows
        if time.time() >= deadline:
            return None, rows
        time.sleep(0.15)


def load_transcript(transcript_path):
    rows = []
    try:
        with open(transcript_path) as fh:
            for line in fh:
                line = line.strip()
                if not line:
                    continue
                try:
                    rows.append(json.loads(line))
                except json.JSONDecodeError:
                    continue
    except Exception:
        pass
    return rows


def _text_blocks(row):
    msg = row.get("message") or {}
    content = msg.get("content")
    if isinstance(content, str):
        return [content]
    if isinstance(content, list):
        return [b.get("text", "") for b in content
                if isinstance(b, dict) and b.get("type") == "text"]
    return []


def is_human_prompt(row):
    if row.get("type") != "user" or row.get("isSidechain"):
        return False
    if row.get("isMeta") or row.get("toolUseResult") is not None:
        return False
    msg = row.get("message") or {}
    content = msg.get("content")
    if isinstance(content, list) and any(
        isinstance(b, dict) and b.get("type") == "tool_result" for b in content
    ):
        return False
    return bool(_text_blocks(row))


def latest_model(rows):
    for row in reversed(rows):
        if row.get("type") == "assistant" and not row.get("isSidechain"):
            m = (row.get("message") or {}).get("model")
            if m:
                return m
    return None


def final_response(rows):
    """The trailing run of assistant text for the most recent turn.

    Walks backwards from the end of the transcript, collecting assistant text
    blocks, and stops at the first thing that is not one (a tool_use, a
    tool_result, or the user prompt). That run IS the final response; anything
    earlier in the turn is an intermediate step and is deliberately dropped.
    """
    chunks, uuid, ts, model = [], None, None, None
    for row in reversed(rows):
        if row.get("isSidechain"):
            continue
        rtype = row.get("type")
        if rtype in ("assistant",):
            msg = row.get("message") or {}
            content = msg.get("content")
            kinds = [b.get("type") for b in content if isinstance(b, dict)] if isinstance(content, list) else []
            texts = _text_blocks(row)
            if texts and all(k == "text" for k in kinds):
                chunks.insert(0, "\n\n".join(t for t in texts if t.strip()))
                if uuid is None:
                    uuid, ts, model = row.get("uuid"), row.get("timestamp"), msg.get("model")
                continue
            break  # tool_use or thinking-only block ends the trailing text run
        if rtype in ("user", "system"):
            break
        # queue-operation / attachment / file-history-snapshot etc: skip over
    text = "\n\n".join(c for c in chunks if c.strip()).strip()
    return text, uuid, ts, model


def normalize_ts(ts):
    if not ts:
        return now_iso()
    try:
        dt = datetime.datetime.fromisoformat(ts.replace("Z", "+00:00"))
        dt = dt.astimezone(datetime.timezone.utc)
        return dt.strftime("%Y-%m-%dT%H:%M:%S.") + f"{dt.microsecond // 1000:03d}Z"
    except Exception:
        return ts
