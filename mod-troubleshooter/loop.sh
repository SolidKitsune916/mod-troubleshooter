#!/bin/bash
set -e

MAX_ITERATIONS=${1:-20}
ITERATION=0
CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || echo "main")

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 Starting Ralph - Mod Troubleshooter"
echo "Branch: $CURRENT_BRANCH"
echo "Max iterations: $MAX_ITERATIONS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Verify PROMPT.md exists
if [ ! -f "PROMPT.md" ]; then
    echo "Error: PROMPT.md not found"
    exit 1
fi

# Verify claude CLI is available
if ! command -v claude &> /dev/null; then
    echo "Error: claude CLI not found. Install Claude Code first."
    exit 1
fi

while [ $ITERATION -lt $MAX_ITERATIONS ]; do
    ITERATION=$((ITERATION + 1))
    echo ""
    echo "═══════════════════════════════════════════"
    echo "  Iteration $ITERATION of $MAX_ITERATIONS"
    echo "═══════════════════════════════════════════"
    echo ""
    
    # Run Claude with the prompt
    cat PROMPT.md | claude --dangerously-skip-permissions 2>&1 | tee -a ralph.log
    
    # Push changes after each iteration
    git push origin "$CURRENT_BRANCH" 2>/dev/null || {
        echo "Push failed or no remote. Continuing..."
    }
    
    # Brief pause between iterations
    sleep 2
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "⚠️  Reached max iterations: $MAX_ITERATIONS"
echo "Check IMPLEMENTATION_PLAN.md for status"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
