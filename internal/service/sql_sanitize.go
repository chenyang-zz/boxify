// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package service

import (
	"strings"
	"unicode"

	"github.com/chenyang-zz/boxify/internal/connection"
)

// sanitizeSQLForPgLike 对于 PostgreSQL 类数据库，转义 LIKE 查询中的特殊字符
func sanitizeSQLForPgLike(dbType connection.ConnectionType, query string) string {
	switch dbType {
	case connection.ConnectionTypePostgreSQL, connection.ConnectionTypeKingbase, connection.ConnectionTypeHighGo, connection.ConnectionTypeVastBase:
		// 有些情况下会出现多层重复引用（例如 """"schema"""" 或 ""schema"""），单次修复不一定收敛。
		// 这里做有限次数的迭代，直到输出不再变化。
		out := query
		for i := 0; i < 3; i++ {
			fixed := fixBrokenDoubleDoubleQuotedIdent(out)
			if fixed == out {
				break
			}
			out = fixed
		}
		return out
	default:
		return query
	}
}

// fixBrokenDoubleDoubleQuotedIdent 修复 PostgreSQL 类数据库中错误的双重双引号标识符
func fixBrokenDoubleDoubleQuotedIdent(query string) string {
	if !strings.Contains(query, `""`) {
		return query
	}

	var b strings.Builder
	b.Grow(len(query))

	inSingle := false
	inDoubleIdent := false
	inLineComment := false
	inBlockComment := false
	dollarTag := ""

	for i := 0; i < len(query); i++ {
		ch := query[i]
		next := byte(0)
		if i+1 < len(query) {
			next = query[i+1]
		}

		if inLineComment {
			b.WriteByte(ch)
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			b.WriteByte(ch)
			if ch == '*' && next == '/' {
				b.WriteByte('/')
				i++
				inBlockComment = false
			}
			continue
		}
		if dollarTag != "" {
			if strings.HasPrefix(query[i:], dollarTag) {
				b.WriteString(dollarTag)
				i += len(dollarTag) - 1
				dollarTag = ""
				continue
			}
			b.WriteByte(ch)
			continue
		}
		if inSingle {
			b.WriteByte(ch)
			if ch == '\'' {
				// escaped single quote
				if next == '\'' {
					b.WriteByte('\'')
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDoubleIdent {
			b.WriteByte(ch)
			if ch == '"' {
				// escaped quote inside identifier
				if next == '"' {
					b.WriteByte('"')
					i++
					continue
				}
				inDoubleIdent = false
			}
			continue
		}

		// --- Outside of all string/comment blocks ---
		if ch == '-' && next == '-' {
			b.WriteByte(ch)
			b.WriteByte('-')
			i++
			inLineComment = true
			continue
		}
		if ch == '/' && next == '*' {
			b.WriteByte(ch)
			b.WriteByte('*')
			i++
			inBlockComment = true
			continue
		}
		if ch == '\'' {
			b.WriteByte(ch)
			inSingle = true
			continue
		}
		if ch == '$' {
			if tag := parseDollarTag(query[i:]); tag != "" {
				b.WriteString(tag)
				i += len(tag) - 1
				dollarTag = tag
				continue
			}
		}

		if ch == '"' {
			// Fix: ""ident"" -> "ident" (only when it looks like a plain identifier)
			// Also handle variants like ""ident""" / """"ident"""" (extra quotes at either side).
			if next == '"' {
				if replacement, advance, ok := tryFixDoubleDoubleQuotedIdent(query, i); ok {
					b.WriteString(replacement)
					i = advance - 1
					continue
				}
			}

			b.WriteByte(ch)
			inDoubleIdent = true
			continue
		}

		b.WriteByte(ch)
	}

	return b.String()
}

// tryFixDoubleDoubleQuotedIdent 尝试修复从 start 位置开始的双重双引号标识符
func tryFixDoubleDoubleQuotedIdent(query string, start int) (replacement string, advance int, ok bool) {
	// start points at the first quote of a broken identifier, usually like:
	//   ""ident""  / ""ident""" / """"ident""""
	if start < 0 || start+1 >= len(query) {
		return "", 0, false
	}
	if query[start] != '"' || query[start+1] != '"' {
		return "", 0, false
	}
	if start > 0 && query[start-1] == '"' {
		return "", 0, false
	}

	runLen := 0
	for start+runLen < len(query) && query[start+runLen] == '"' {
		runLen++
	}
	if runLen < 2 || runLen%2 == 1 {
		// Odd run (e.g. """...) can be a valid quoted identifier with escaped quotes.
		return "", 0, false
	}

	contentStart := start + runLen
	j := contentStart
	for j < len(query) {
		if query[j] == '"' {
			endRunLen := 0
			for j+endRunLen < len(query) && query[j+endRunLen] == '"' {
				endRunLen++
			}
			if endRunLen >= 2 {
				content := strings.TrimSpace(query[contentStart:j])
				if looksLikeIdentifierContent(content) {
					return `"` + content + `"`, j + endRunLen, true
				}
				return "", 0, false
			}
		}
		// Fast abort: identifier-like content should not span lines.
		if query[j] == '\n' || query[j] == '\r' {
			break
		}
		j++
	}
	return "", 0, false
}

// looksLikeIdentifierContent 判断字符串是否看起来像标识符内容
func looksLikeIdentifierContent(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	for _, r := range s {
		if r == '_' || r == '$' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

// parseDollarTag 尝试解析 PostgreSQL 风格的 dollar-quoted tag
func parseDollarTag(s string) string {
	// Match: $tag$ where tag is [A-Za-z0-9_]* (can be empty => $$)
	if len(s) < 2 || s[0] != '$' {
		return ""
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c == '$' {
			return s[:i+1]
		}
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return ""
		}
	}
	return ""
}
