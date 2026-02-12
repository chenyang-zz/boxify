// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import "strings"

// quoteIdentByType 根据数据库类型对标识符进行适当的引用和转义，防止SQL注入和语法错误
func quoteIdentByType(dbType string, ident string) string {
	if ident == "" {
		return ident
	}

	switch dbType {
	case "mysql", "mariadb", "tdengine":
		return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
	case "sqlserver":
		escaped := strings.ReplaceAll(ident, "]", "]]")
		return "[" + escaped + "]"
	default:
		return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
	}
}
