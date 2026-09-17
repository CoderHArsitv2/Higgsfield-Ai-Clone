#!/usr/bin/env python3
"""UserPromptSubmit hook: append the verbatim prompt to .agent-logs/.

Writes nothing to stdout -- stdout from a UserPromptSubmit hook is injected
into the model's context, which would contaminate the very thing being logged.
"""
import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import agentlog as L  # noqa: E402


def main(raw):
    payload = json.loads(raw)
    L.debug_dump("UserPromptSubmit", payload)

    session_id = payload.get("session_id")
    prompt = payload.get("prompt")
    if not session_id or prompt is None:
        return

    # The model is read from the transcript's most recent assistant message.
    # On the first turn of a session there isn't one yet, so it is marked
    # pending and the Stop hook fills it in for that same turn.
    rows = L.load_transcript(payload.get("transcript_path", ""))
    model = L.latest_model(rows) or "pending"

    num = L.next_prompt_num(session_id)
    L.append_entry(session_id, "PROMPT", num, L.now_iso(), model, prompt)

    st = L.read_state(session_id)
    st["open_prompt_num"] = num
    st["open_prompt_id"] = payload.get("prompt_id")
    L.write_state(session_id, st)


if __name__ == "__main__":
    raw = sys.stdin.read()
    try:
        main(raw)
    except Exception:
        import traceback
        L.record_failure("UserPromptSubmit", raw, traceback.format_exc())
    sys.exit(0)  # a broken hook must never block the session
