#!/usr/bin/env python3
"""Stop hook: append the final assistant response for the turn.

The response text comes from `last_assistant_message` on the hook payload.
That is the turn's final text, already isolated by Claude Code -- no thinking,
no tool calls, no intermediate steps -- and it avoids a race: the Stop hook can
fire fractionally before the message is flushed to the transcript JSONL.
Transcript parsing remains as a fallback only.
"""
import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import agentlog as L  # noqa: E402


def main(raw):
    payload = json.loads(raw)
    L.debug_dump("Stop", payload)

    session_id = payload.get("session_id")
    transcript = payload.get("transcript_path", "")
    prompt_id = payload.get("prompt_id")
    if not session_id:
        return

    st = L.read_state(session_id)
    if prompt_id and st.get("logged_prompt_id") == prompt_id:
        return  # this turn's response is already on record

    model, rows = L.wait_for_model(transcript)

    text = (payload.get("last_assistant_message") or "").strip()
    if not text:
        text, _uuid, _ts, fb_model = L.final_response(rows)
        model = model or fb_model

    # Self-heal: if this turn's prompt was never captured (hook installed
    # mid-session, or a resumed session) recover it from the transcript so the
    # log never carries an orphaned response.
    if L.session_file(session_id) is None:
        for row in reversed(rows):
            if L.is_human_prompt(row):
                content = (row.get("message") or {}).get("content")
                ptext = "\n\n".join(
                    b.get("text", "") for b in content
                    if isinstance(b, dict) and b.get("type") == "text"
                ) if isinstance(content, list) else str(content)
                L.append_entry(session_id, "PROMPT", 1,
                               L.normalize_ts(row.get("timestamp")),
                               model or "unknown", ptext)
                break

    num = st.get("open_prompt_num") if st.get("open_prompt_id") == prompt_id else None
    if num is None:
        num = L.current_prompt_num(session_id)

    L.resolve_pending_model(session_id, num, model)
    L.append_entry(
        session_id, "RESPONSE", num, L.now_iso(), model or "unknown",
        text if text else "[no final text response - turn ended without assistant text]",
    )

    st["logged_prompt_id"] = prompt_id
    L.write_state(session_id, st)


if __name__ == "__main__":
    raw = sys.stdin.read()
    try:
        main(raw)
    except Exception:
        import traceback
        L.record_failure("Stop", raw, traceback.format_exc())
    sys.exit(0)  # a broken hook must never block the session
