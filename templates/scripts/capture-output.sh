#!/bin/bash
# Captures PostToolUse output and injects it as additional context for the workflow.
# Used by Claude Code plugin hooks to implement the equivalent of Antigravity's // capture annotation.
INPUT=$(cat)
OUTPUT=$(echo "$INPUT" | jq -r '.tool_result // empty' 2>/dev/null)
if [ -n "$OUTPUT" ]; then
    echo "{\"additionalContext\": \"Captured output: $OUTPUT\"}"
fi