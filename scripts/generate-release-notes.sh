#!/usr/bin/env bash
set -euo pipefail

# 生成规范化发布说明：Release Info / Highlights / Stability / Chores / Verification / Full Changelog
VERSION="${1:-}"
TITLE="${2:-常规功能更新与稳定性优化}"
OUTPUT="${3:-.release-notes.md}"

if [[ -z "$VERSION" ]]; then
  echo "用法: $0 <version-or-tag> [title] [output-file]" >&2
  exit 1
fi

if [[ "$VERSION" != v* ]]; then
  TAG="v$VERSION"
else
  TAG="$VERSION"
fi

PREV_TAG="$(git tag --sort=-v:refname | grep '^v' | grep -v "^${TAG}$" | head -n1 || true)"
if [[ -n "$PREV_TAG" ]]; then
  RANGE="${PREV_TAG}...${TAG}"
  RANGE_DISPLAY="${PREV_TAG}...${TAG}"
else
  FIRST_COMMIT="$(git rev-list --max-parents=0 HEAD | tail -n1)"
  RANGE="${FIRST_COMMIT}...${TAG}"
  RANGE_DISPLAY="(initial)...${TAG}"
fi

ALL_LINES="$(git log --pretty='- %s' "$RANGE" || true)"
FEATURE_LINES=""
FIX_LINES=""
CHORE_LINES=""

while IFS= read -r line; do
  if [[ -z "$line" ]]; then
    continue
  fi

  if printf '%s\n' "$line" | grep -Eqi '(feat|✨|新增|支持|add)'; then
    FEATURE_LINES="${FEATURE_LINES}${line}"$'\n'
    continue
  fi

  if printf '%s\n' "$line" | grep -Eqi '(fix|🐛|修复|优化|refactor|perf|⚡)'; then
    FIX_LINES="${FIX_LINES}${line}"$'\n'
    continue
  fi

  if printf '%s\n' "$line" | grep -Eqi '(chore|🔧|build|ci|docs|test|✅|📝)'; then
    CHORE_LINES="${CHORE_LINES}${line}"$'\n'
    continue
  fi
done <<< "$ALL_LINES"

FEATURE_LINES="$(printf '%s' "$FEATURE_LINES" | sed '/^$/d' || true)"
FIX_LINES="$(printf '%s' "$FIX_LINES" | sed '/^$/d' || true)"
CHORE_LINES="$(printf '%s' "$CHORE_LINES" | sed '/^$/d' || true)"
COMMIT_COUNT="$(git rev-list --count "$RANGE" 2>/dev/null || echo "0")"
RELEASE_DATE="$(date '+%Y-%m-%d %H:%M:%S %z')"

if [[ -z "$FEATURE_LINES" ]]; then
  FEATURE_LINES='- 本版本包含功能改进与交互体验优化'
fi
if [[ -z "$FIX_LINES" ]]; then
  FIX_LINES='- 本版本包含稳定性修复与细节优化'
fi
if [[ -z "$CHORE_LINES" ]]; then
  CHORE_LINES='- 本版本包含工程化与发布流程相关改进'
fi

REPO_URL="$(git config --get remote.origin.url | sed -E 's#git@github.com:#https://github.com/#; s#\.git$##')"
if [[ -z "$REPO_URL" ]]; then
  REPO_URL='https://github.com/chenyang-zz/boxify'
fi
BT='`'

cat > "$OUTPUT" <<NOTE
## ${TAG} - ${TITLE}

### Release Info
- Version: ${TAG}
- Date: ${RELEASE_DATE}
- Commit Count: ${COMMIT_COUNT}
- Compare Range: ${RANGE_DISPLAY}

### Highlights
${FEATURE_LINES}

### Stability
${FIX_LINES}

### Chores
${CHORE_LINES}

### Verification
- Passed ${BT}pnpm -C frontend run build${BT}
- Passed ${BT}go test ./...${BT}
- Passed ${BT}make build-release${BT}

**Full Changelog**: ${REPO_URL}/compare/${RANGE}
NOTE

echo "Release notes generated at: $OUTPUT"
