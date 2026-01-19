// file: url_match.go
// Package urlmatch 提供与 Gin 路由语义一致的 URL 匹配工具（对称匹配）
package utils

import (
	"net/url"
	"strings"
)

// UrlIsEq 判断两个 URL / 路由定义是否等价
//
// 支持：
//   - 固定路径       /api/v1/users
//   - 参数路径       /api/v1/users/:id
//   - Gin 通配路径   /api/v1/users/*path
//   - 忽略 query     ?a=1
//
// source / target 顺序无关
func UrlIsEq(source, target string) bool {
	sp := normalizePath(extractPath(source))
	tp := normalizePath(extractPath(target))

	ss := splitPath(sp)
	ts := splitPath(tp)

	i := 0
	for {
		// 任意一方到达末尾
		if i >= len(ss) || i >= len(ts) {
			break
		}

		s := ss[i]
		t := ts[i]

		// 任意一方是 Gin 通配符 *xxx
		if isWildcardSegment(s) || isWildcardSegment(t) {
			return true
		}

		// 单段参数 :id
		if isParamSegment(s) || isParamSegment(t) {
			i++
			continue
		}

		// 固定段必须相等
		if s != t {
			return false
		}

		i++
	}

	// 处理剩余段：
	// 如果剩余的一方只有一个 *path，则仍然匹配
	if i < len(ss) && isWildcardSegment(ss[i]) {
		return true
	}
	if i < len(ts) && isWildcardSegment(ts[i]) {
		return true
	}

	// 否则，必须同时耗尽
	return i == len(ss) && i == len(ts)
}

// ======================== helpers ========================

func extractPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		if idx := strings.Index(rawURL, "?"); idx > -1 {
			return rawURL[:idx]
		}
		return rawURL
	}
	return u.Path
}

func normalizePath(p string) string {
	if p == "" {
		return "/"
	}

	if idx := strings.Index(p, "?"); idx > -1 {
		p = p[:idx]
	}

	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	if len(p) > 1 {
		p = strings.TrimRight(p, "/")
	}

	return p
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

func isParamSegment(seg string) bool {
	return strings.HasPrefix(seg, ":")
}

func isWildcardSegment(seg string) bool {
	return strings.HasPrefix(seg, "*")
}
