// Copyright Contributors to the KubeOpenCode project

package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRewriteHTMLResponse(t *testing.T) {
	proxyBase := "/api/v1/namespaces/test/agents/my-agent/web"

	tests := []struct {
		name     string
		inputHTML string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "rewrites script src paths",
			inputHTML: `<html><head><script src="/assets/app.js"></script></head><body></body></html>`,
			wantContains: []string{
				`src="` + proxyBase + `/assets/app.js"`,
			},
			wantNotContains: []string{
				`src="/assets/app.js"`,
			},
		},
		{
			name: "rewrites link href paths",
			inputHTML: `<html><head><link href="/assets/style.css" rel="stylesheet"></head><body></body></html>`,
			wantContains: []string{
				`href="` + proxyBase + `/assets/style.css"`,
			},
			wantNotContains: []string{
				`href="/assets/style.css"`,
			},
		},
		{
			name: "injects fetch monkey-patch script",
			inputHTML: `<html><head></head><body></body></html>`,
			wantContains: []string{
				`<script>`,
				`window.fetch=function`,
				proxyBase,
				`</script></head>`,
			},
		},
		{
			name: "handles multiple asset paths",
			inputHTML: `<html><head><script src="/assets/a.js"></script><link href="/assets/b.css"></head><body><img src="/favicon.ico"></body></html>`,
			wantContains: []string{
				`src="` + proxyBase + `/assets/a.js"`,
				`href="` + proxyBase + `/assets/b.css"`,
				`src="` + proxyBase + `/favicon.ico"`,
			},
		},
		{
			name: "does not rewrite relative paths",
			inputHTML: `<html><head><script src="relative.js"></script></head><body></body></html>`,
			wantContains: []string{
				`src="relative.js"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: http.Header{"Content-Type": []string{"text/html"}},
				Body:   io.NopCloser(bytes.NewBufferString(tt.inputHTML)),
			}

			if err := rewriteHTMLResponse(resp, proxyBase); err != nil {
				t.Fatalf("rewriteHTMLResponse() error = %v", err)
			}

			body, _ := io.ReadAll(resp.Body)
			result := string(body)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result does not contain %q\n\nGot:\n%s", want, result)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(result, notWant) {
					t.Errorf("result should not contain %q\n\nGot:\n%s", notWant, result)
				}
			}

			// Verify Content-Length is updated
			if resp.ContentLength != int64(len(body)) {
				t.Errorf("ContentLength = %d, want %d", resp.ContentLength, len(body))
			}
		})
	}
}

func TestIsOpenCodeAPIPath(t *testing.T) {
	apiPaths := []string{
		"/session/list", "/global/health", "/global/event",
		"/event", "/permission/123/reply", "/file/content",
		"/pty", "/config/providers", "/path", "/vcs",
	}
	for _, p := range apiPaths {
		if !isOpenCodeAPIPath(p) {
			t.Errorf("expected %q to be an API path", p)
		}
	}

	nonAPIPaths := []string{
		"/", "/assets/app.js", "/workspace/session/123",
		"/some-spa-route", "/favicon.ico",
	}
	for _, p := range nonAPIPaths {
		if isOpenCodeAPIPath(p) {
			t.Errorf("expected %q to NOT be an API path", p)
		}
	}
}

func TestBuildFetchPatchScript(t *testing.T) {
	script := buildFetchPatchScript("/api/v1/namespaces/ns/agents/ag/web")

	if !strings.HasPrefix(script, "<script>") {
		t.Error("script should start with <script>")
	}
	if !strings.HasSuffix(script, "</script>") {
		t.Error("script should end with </script>")
	}
	if !strings.Contains(script, "/api/v1/namespaces/ns/agents/ag/web") {
		t.Error("script should contain the proxy base path")
	}
	if !strings.Contains(script, "window.fetch") {
		t.Error("script should patch window.fetch")
	}
}
