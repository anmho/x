package tools

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceReadListAndSearch(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "README.md"), "# Project X\nhello tools\n")
	mustWriteFile(t, filepath.Join(root, "nested", "guide.txt"), "hello from nested\n")
	reg := NewRegistry(root)

	readRaw, isError, err := reg.Call("workspace_read_file", mustJSON(map[string]any{
		"path":       "README.md",
		"start_line": 2,
		"end_line":   2,
	}))
	if err != nil {
		t.Fatalf("workspace_read_file err: %v", err)
	}
	if isError {
		t.Fatalf("workspace_read_file isError: %s", readRaw)
	}
	var read workspaceReadResult
	if err := json.Unmarshal([]byte(readRaw), &read); err != nil {
		t.Fatalf("unmarshal read result: %v", err)
	}
	if read.Content != "hello tools" {
		t.Fatalf("unexpected read content: %#v", read)
	}

	listRaw, isError, err := reg.Call("workspace_list_files", mustJSON(map[string]any{
		"path":      ".",
		"recursive": true,
	}))
	if err != nil {
		t.Fatalf("workspace_list_files err: %v", err)
	}
	if isError {
		t.Fatalf("workspace_list_files isError: %s", listRaw)
	}
	var listed workspaceListResult
	if err := json.Unmarshal([]byte(listRaw), &listed); err != nil {
		t.Fatalf("unmarshal list result: %v", err)
	}
	if len(listed.Paths) != 2 || listed.Paths[0] != "README.md" {
		t.Fatalf("unexpected listed paths: %#v", listed.Paths)
	}

	searchRaw, isError, err := reg.Call("workspace_search", mustJSON(map[string]any{
		"query": "hello",
	}))
	if err != nil {
		t.Fatalf("workspace_search err: %v", err)
	}
	if isError {
		t.Fatalf("workspace_search isError: %s", searchRaw)
	}
	var search workspaceSearchResult
	if err := json.Unmarshal([]byte(searchRaw), &search); err != nil {
		t.Fatalf("unmarshal search result: %v", err)
	}
	if len(search.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %#v", search.Matches)
	}
}

func TestWorkspaceApplyPatchAndGit(t *testing.T) {
	root := initGitRepo(t)
	mustWriteFile(t, filepath.Join(root, "notes.txt"), "before\n")
	runGit(t, root, "add", "notes.txt")
	runGit(t, root, "commit", "-m", "init")

	reg := NewRegistry(root)

	patch := strings.Join([]string{
		"--- a/notes.txt",
		"+++ b/notes.txt",
		"@@ -1 +1 @@",
		"-before",
		"+after",
		"",
	}, "\n")
	patchRaw, isError, err := reg.Call("workspace_apply_patch", mustJSON(map[string]any{"patch": patch}))
	if err != nil {
		t.Fatalf("workspace_apply_patch err: %v", err)
	}
	if isError {
		t.Fatalf("workspace_apply_patch isError: %s", patchRaw)
	}
	if got := string(mustReadFile(t, filepath.Join(root, "notes.txt"))); !strings.Contains(got, "after") {
		t.Fatalf("patch did not modify file: %q", got)
	}

	diffRaw, isError, err := reg.Call("git_diff", mustJSON(map[string]any{}))
	if err != nil {
		t.Fatalf("git_diff err: %v", err)
	}
	if isError {
		t.Fatalf("git_diff isError: %s", diffRaw)
	}
	if !strings.Contains(diffRaw, "+after") {
		t.Fatalf("expected git diff output, got %q", diffRaw)
	}

	statusRaw, isError, err := reg.Call("git_status", mustJSON(map[string]any{}))
	if err != nil {
		t.Fatalf("git_status err: %v", err)
	}
	if isError {
		t.Fatalf("git_status isError: %s", statusRaw)
	}
	if !strings.Contains(statusRaw, "notes.txt") {
		t.Fatalf("expected modified file in status, got %q", statusRaw)
	}

	checkoutRaw, isError, err := reg.Call("git_checkout", mustJSON(map[string]any{
		"ref":           "feature/runtime-tools",
		"create_branch": true,
	}))
	if err != nil {
		t.Fatalf("git_checkout err: %v", err)
	}
	if isError {
		t.Fatalf("git_checkout isError: %s", checkoutRaw)
	}
	if !strings.Contains(checkoutRaw, "feature/runtime-tools") {
		t.Fatalf("expected branch checkout output, got %q", checkoutRaw)
	}
}

func TestWorkspaceToolsRejectPathEscape(t *testing.T) {
	reg := NewRegistry(t.TempDir())

	raw, isError, err := reg.Call("workspace_read_file", mustJSON(map[string]any{
		"path": "../secret.txt",
	}))
	if err != nil {
		t.Fatalf("workspace_read_file err: %v", err)
	}
	if !isError {
		t.Fatalf("expected path escape to be rejected")
	}
	if !strings.Contains(raw, "escapes repo root") {
		t.Fatalf("unexpected error: %q", raw)
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.name", "Codex Test")
	runGit(t, root, "config", "user.email", "codex@example.com")
	return root
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
