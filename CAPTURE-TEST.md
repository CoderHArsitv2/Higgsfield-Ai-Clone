# CAPTURE-TEST.md

Proof that automatic prompt/response capture is installed and firing, per step 4 of
the 8x assignment brief. Written before any assignment code was built.

## 1. Setup

| | |
| --- | --- |
| **Tool** | Claude Code `2.1.228`, running as the native VS Code extension |
| **Model** | `claude-opus-5` — exact id `claude-opus-5[1m]` (the 1M-context variant). The transcript records it as `claude-opus-5`, and that is what the logs show. |
| **Planner / executor** | The same model does both. There is no separate planning model. A mid-build switch *would* be visible: the model is read per turn from the transcript's assistant message rather than hardcoded once, so it is recorded per `LOG_ENTRY`, not just in the frontmatter. |
| **Automatic mechanism?** | **Yes.** Claude Code has lifecycle hooks configured in `.claude/settings.json`. I confirmed the event names and payload schema by dumping raw hook stdin rather than trusting documentation from memory. |

## 2. Mechanism

Two lifecycle hooks. Nothing is run by hand, and nothing depends on me remembering to
do anything.

**Config file changed:** [`.claude/settings.json`](.claude/settings.json)

| Event | Script | What it does |
| --- | --- | --- |
| `UserPromptSubmit` | [`.claude/hooks/capture_prompt.py`](.claude/hooks/capture_prompt.py) | Receives the prompt verbatim on stdin as JSON and appends a `PROMPT` entry immediately — before the model runs, so an interrupted or crashed turn still leaves the prompt on record. |
| `Stop` | [`.claude/hooks/capture_response.py`](.claude/hooks/capture_response.py) | Fires at end of turn. Appends the `RESPONSE` entry. |

Shared logic lives in [`.claude/hooks/agentlog.py`](.claude/hooks/agentlog.py).

**What gets captured — prompt and final response only.** The `Stop` payload carries a
`last_assistant_message` field, which is the turn's final assistant text already
isolated by Claude Code. Thinking blocks, tool calls, tool results, file reads, diffs
and retries are all structurally absent from it — they are never read, so they cannot
leak in. Parsing the transcript JSONL is kept only as a fallback, and that fallback
walks backwards from the end of the turn collecting the trailing run of assistant
`text` blocks, stopping at the first `tool_use` / `thinking` / `tool_result`.

**Design decisions worth stating:**

- **Prompt and response are written by different hooks at different times**, paired via
  the `prompt_id` that both payloads share. Dedupe is on `prompt_id`, so a repeated
  `Stop` cannot double-log a turn.
- **A hook failure is never silent.** Both hooks always `exit 0` — a capture bug must
  never block a session — but any exception appends the traceback *and the raw stdin*
  to `.agent-logs/_capture-errors.log`. A gap in the log is therefore always
  explainable. That file does not currently exist, which is the pass condition.
- **`capture_prompt.py` writes nothing to stdout**, because stdout from a
  `UserPromptSubmit` hook is injected into the model's context and would contaminate
  the very thing being logged.
- **Entry text is never rewritten.** Only the YAML frontmatter is regenerated on each
  append, since `total_exchanges` / `last_prompt_time` are by definition not knowable
  when the file is created. The one exception is documented in section 4.
- **`.agent-logs/` is not in `.gitignore`.** The `.gitignore` says so explicitly.

## 3. Where the canaries landed

Two canaries, in two separate sessions, neither of which was the session that installed
the hooks. Both were launched headless (`claude -p`) so each got its own fresh session
id and its own file — proving the hook is installed repo-wide, not just in the session
that wrote it.

- Canary 1 → [`.agent-logs/2026-09-17_11-27-24_67d3a9e0-5fa1-40dc-97cc-4aff922aea4c.md`](.agent-logs/2026-09-17_11-27-24_67d3a9e0-5fa1-40dc-97cc-4aff922aea4c.md)
- Canary 2 → [`.agent-logs/2026-09-17_11-28-06_74cf03ac-6d07-412f-b4d9-92685f49a00e.md`](.agent-logs/2026-09-17_11-28-06_74cf03ac-6d07-412f-b4d9-92685f49a00e.md)

### Canary 1, pasted raw

```
[LOG_ENTRY type=PROMPT num=1 session=67d3a9e0]
timestamp: 2026-09-17T11:27:24.561Z
model: claude-opus-5

CAPTURE TEST — 8x assignment, Harshit Saxena


[LOG_ENTRY type=RESPONSE num=1 session=67d3a9e0]
timestamp: 2026-09-17T11:27:59.107Z
model: claude-opus-5

This is a capture-hook test harness, not a code task — so here's the state of the pipeline plus this turn's capture marker.

**Setup**
- `.claude/settings.json` wires two hooks: `UserPromptSubmit` → `capture_prompt.py`, `Stop` → `capture_response.py`, both writing into `.agent-logs/`.
- `capture_response.py` pulls the turn's final text from `last_assistant_message` (transcript JSONL parsing is fallback only), dedupes on `prompt_id`, self-heals a missing PROMPT entry, and always `exit(0)` so a hook bug can't block the session.

**Capture status across sessions**

| Session | Prompts | Responses | Model recorded |
|---|---|---|---|
| `744620c7` | 1 | 1 | unknown |
| `467d04de` | 1 | 1 | unknown |
| `8d4b701f` | 1 | 1 | unknown |
| `4d79eab0` | 2 | 1 | claude-opus-5 |
| `67d3a9e0` (this one) | 1 | 0 — pending this turn | pending |

Two things worth flagging:

1. **Model resolution is unreliable.** Only `4d79eab0` resolved to `claude-opus-5`; the other three logged `unknown` for the whole file. `wait_for_model(transcript)` is apparently losing the race against the transcript flush in most sessions. The `resolve_pending_model` path exists for exactly this, but it only helps when the model eventually shows up.

2. **`4d79eab0` has an orphaned prompt** — PROMPT num=2 with no matching RESPONSE. That's the expected shape if the session was interrupted before Stop fired (Stop doesn't run on abort), so it's likely not a bug, but it's the one asymmetry in the logs.

This turn's response will be appended to `.agent-logs/2026-09-17_11-27-24_67d3a9e0-...md` as `RESPONSE num=1` when the Stop hook fires — re-read that file after this turn to confirm the capture, and check whether `model:` resolves to `claude-opus-5[1m]` or falls back to `unknown` again.

I didn't change anything. If you want, I can make model resolution deterministic by reading `CLAUDE_*` env/payload fields instead of racing the transcript.
```

### Canary 2, pasted raw

```
[LOG_ENTRY type=PROMPT num=1 session=74cf03ac]
timestamp: 2026-09-17T11:28:06.427Z
model: claude-opus-5

CAPTURE TEST — 8x assignment, Harshit Saxena (second session)


[LOG_ENTRY type=RESPONSE num=1 session=74cf03ac]
timestamp: 2026-09-17T11:30:26.321Z
model: claude-opus-5

Received — second session, capture confirmed.

Beyond the confirmation, I checked the prior session's (`67d3a9e0`) two flags, since it offered to "fix" one of them.

**Capture pipeline is working.** Your prompt landed in `.agent-logs/2026-09-17_11-28-06_74cf03ac-...md` at `11:28:06` with `model: pending`; `resolve_pending_model` should rewrite that to the real model when the Stop hook fires on this turn.

**The "model resolution is unreliable" flag is probably a false alarm.** That conclusion came from a table comparing all five logs — but those logs weren't written by the same code. The four `unknown` sessions contain the placeholder string `turn ended on a tool call`, and that string does not exist in the current `capture_response.py:62`, which writes `turn ended without assistant text` instead. So those four predate the hook you're running now, and comparing them against `67d3a9e0` mixes versions. The one log written by the current hook resolved `claude-opus-5` correctly. I'd hold off on rewriting `wait_for_model` until a session fails under the *current* code.

I couldn't confirm this from the transcripts directly — reads and greps outside the working directory are blocked here, and Bash approval isn't available in a non-interactive session — so that's inference from the placeholder strings, not proof.

**One real defect, from reading the code.** `ENTRY_RE` (`agentlog.py:23`) is matched against the log body with no escaping of entry text, and `parse_entries` feeds `build_header`. So if any *response text* contains a line matching the entry-header pattern followed by `timestamp:` and `model:` lines, it gets parsed as a real entry — inflating `total_exchanges`, and potentially hijacking `first_prompt_time`/`model` in the frontmatter. This project is the worst case for it: discussing the log format is the task, so a response that quotes a header block verbatim corrupts the log that stores it. I deliberately avoided emitting that pattern above. Cheap fix is to indent or fence entry text on write, or anchor `ENTRY_RE` to only match at known entry offsets.

Minor, while I was in there: `now_iso` (`agentlog.py:31`) calls `datetime.now()` twice and takes seconds from the first and milliseconds from the second, so it can emit a stray off-by-one-millisecond timestamp.

Separately — the MCP servers for Canva, HubSpot, Linear, Lucid, Notion, and Slack aren't authorized, and I can't run OAuth from a non-interactive session. Authorize them in your claude.ai connector settings if you need them.

I changed nothing in the repo. Say the word and I'll patch the `ENTRY_RE` escaping.
```

## 4. What I tried first that did not work

These are real, and the broken output is still in `.agent-logs/` rather than deleted.

**a. First version parsed the transcript JSONL for the response. It lost a race.**
My first `Stop` hook read the session transcript and extracted the final assistant text
from it. It produced `[no final text response - turn ended on a tool call]` for turns
that plainly had a response. Cause: `Stop` fires fractionally *before* Claude Code
flushes the final assistant message to the transcript file, so the hook parsed a file
that did not contain the answer yet. Dumping the raw hook payload showed the `Stop`
event already hands over `last_assistant_message` directly — the transcript parsing was
never necessary and was the source of the bug. Broken output preserved in
`.agent-logs/2026-09-17_11-23-51_744620c7-*.md` and
`.agent-logs/2026-09-17_11-24-02_467d04de-*.md`.

**b. `model: unknown` on the first turn of every session.**
Same root cause, different symptom. The model was resolved by reading the last
assistant message out of the transcript — but on the *first* turn of a fresh session no
assistant message exists yet. Fixed in two parts: the `Stop` hook now polls briefly for
the model instead of reading once, and `capture_prompt.py` writes `model: pending` on a
first turn which the `Stop` hook then resolves for that same turn. This is the one case
where a written entry is touched after the fact: it replaces the literal string
`pending` in the `model:` field of the turn currently being written. It never touches
prompt or response text. I am flagging it rather than hiding it.

**c. My own test harness was broken, and the hook hid it.**
An early unit test piped a payload built with zsh `echo`, which expands `\n` and `\t`,
producing invalid JSON. The hook caught the parse error and exited 0 — so the prompt
was silently dropped and the test looked like it passed. That is the worst possible
failure mode for a capture system: no output and no complaint. This is what prompted
the `_capture-errors.log` design above.

**d. Hooks are snapshotted at session start.**
The session I was working in when I installed the hooks did not capture its own early
turns — the first two exchanges of `4d79eab0` predate the hook and are not in
`.agent-logs/`. Claude Code picked the settings up partway through, and that session's
later turns are captured. Nothing was backfilled or reconstructed; the gap is real and
is left as a gap.

**e. An orphaned `PROMPT` with no `RESPONSE` is expected, not a bug.**
`Stop` does not fire when a turn is interrupted. `4d79eab0` contains a `PROMPT` with no
matching `RESPONSE` for exactly that reason. Left in place.

**f. Prompts captured in the IDE carry an `<ide_opened_file>` prefix.**
The VS Code extension prepends editor context (e.g. which file the user just opened) to
the prompt before it reaches the `UserPromptSubmit` hook. So a captured prompt is
verbatim what was *submitted*, which is not always identical to what was *typed*. I
considered stripping the tag and decided not to: the brief says no cleanup, and
stripping it would be me deciding what counts as part of the prompt. It is left in, and
flagged here so it is not mistaken for noise. Headless (`claude -p`) prompts have no
such prefix, which is why the canaries are clean.

**g. One entry in this session's own log records the bug being fixed.**
`.agent-logs/2026-09-17_11-25-38_4d79eab0-*.md` has `RESPONSE num=1` reading
`[no final text response - turn ended on a tool call]` — that turn did have a response.
It was written by the pre-fix hook, minutes before the rewrite landed. It is wrong, and
it stays. The fix is visible in the same file: later entries in that session capture
correctly.

**h. The author handle was corrected partway through, and the logs were touched.**
The hooks initially took the author from global git config (`CoderHArsit`); the correct
handle is `coderHArsitv2`. I updated the `author:` field in the YAML frontmatter and the
header line of the already-written log files. That is metadata, and it is regenerated on
every append anyway. **No `LOG_ENTRY` body was edited.** One response entry in
`4d79eab0` still contains the string `CoderHArsit` inside its text, because that is what
was actually said at the time — it was deliberately left alone. The three commits made
before the correction were re-authored via `git filter-branch`; nothing had been pushed.
