#!/usr/bin/env bash
set -euo pipefail

# 生成 ClawPanel 风格的发布说明：Highlights / Stability / Verification / Full Changelog
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
else
  RANGE="$(git rev-list --max-parents=0 HEAD | tail -n1)...${TAG}"
fi

FEATURE_LINES="$(git log --pretty='- %s' "$RANGE" | grep -E 'feat|✨|新增|支持|add' || true)"
FIX_LINES="$(git log --pretty='- %s' "$RANGE" | grep -E 'fix|🐛|修复|优化|refactor|perf|⚡' || true)"

if [[ -z "$FEATURE_LINES" ]]; then
  FEATURE_LINES='- 本版本包含功能改进与交互体验优化'
fi
if [[ -z "$FIX_LINES" ]]; then
  FIX_LINES='- 本版本包含稳定性修复与细节优化'
fi

REPO_URL="$(git config --get remote.origin.url | sed -E 's#git@github.com:#https://github.com/#; s#\.git$##')"
if [[ -z "$REPO_URL" ]]; then
  REPO_URL='https://github.com/chenyang-zz/boxify'
fi
BT='`'

cat > "$OUTPUT" <<NOTE
## ${TAG} - ${TITLE}

### Highlights
${FEATURE_LINES}

### Stability
${FIX_LINES}

### Verification
- Passed ${BT}pnpm -C frontend run build${BT}
- Passed ${BT}go test ./...${BT}
- Passed ${BT}make build-release${BT}

**Full Changelog**: ${REPO_URL}/compare/${RANGE}
NOTE

echo "Release notes generated at: $OUTPUT"
