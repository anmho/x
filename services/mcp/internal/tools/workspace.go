package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type workspaceReadResult struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Content   string `json:"content"`
}

type workspaceListResult struct {
	BasePath string   `json:"base_path"`
	Paths    []string `json:"paths"`
}

type workspaceSearchMatch struct {
	Path       string `json:"path"`
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
}

type workspaceSearchResult struct {
	Query    string                 `json:"query"`
	BasePath string                 `json:"base_path"`
	Matches  []workspaceSearchMatch `json:"matches"`
}

type applyPatchResult struct {
	Applied bool   `json:"applied"`
	Status  string `json:"status"`
}

func workspaceTools() []Tool {
	return []Tool{
		{
			Name:        "workspace_read_file",
			Description: "Read a repo-relative text file, optionally scoped to a line range",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "path": {"type": "string", "description": "Repo-relative file path"},
    "start_line": {"type": "number", "description": "Optional 1-based start line"},
    "end_line": {"type": "number", "description": "Optional 1-based end line"}
  },
  "required": ["path"]
}`,
		},
		{
			Name:        "workspace_list_files",
			Description: "List repo-relative files beneath a directory",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "path": {"type": "string", "description": "Optional repo-relative directory path"},
    "recursive": {"type": "boolean", "description": "Whether to recurse into subdirectories (default true)"},
    "limit": {"type": "number", "description": "Maximum number of file paths to return (default 200)"},
    "include_hidden": {"type": "boolean", "description": "Whether to include dotfiles/directories"}
  },
  "required": []
}`,
		},
		{
			Name:        "workspace_search",
			Description: "Search repo files for a text query and return matching lines",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Text to search for"},
    "path": {"type": "string", "description": "Optional repo-relative directory path"},
    "limit": {"type": "number", "description": "Maximum number of matches to return (default 50)"},
    "case_sensitive": {"type": "boolean", "description": "Whether the search should be case-sensitive"}
  },
  "required": ["query"]
}`,
		},
		{
			Name:        "workspace_apply_patch",
			Description: "Apply a unified diff patch to the repository working tree",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "patch": {"type": "string", "description": "Unified diff patch text"}
  },
  "required": ["patch"]
}`,
		},
		{
			Name:        "git_status",
			Description: "Return git status for the repository working tree",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {},
  "required": []
}`,
		},
		{
			Name:        "git_diff",
			Description: "Return git diff output for the repository or a specific path",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "path": {"type": "string", "description": "Optional repo-relative path to diff"},
    "ref": {"type": "string", "description": "Optional git ref to diff against"},
    "cached": {"type": "boolean", "description": "Whether to diff the index instead of the working tree"}
  },
  "required": []
}`,
		},
		{
			Name:        "git_checkout",
			Description: "Switch the repository to an existing ref or create a new branch",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "ref": {"type": "string", "description": "Branch, tag, or commit to check out"},
    "create_branch": {"type": "boolean", "description": "Create and check out a new branch"}
  },
  "required": ["ref"]
}`,
		},
	}
}

func (r *Registry) callWorkspaceTool(name string, args map[string]interface{}) (string, bool, error) {
	getString := func(key string) string {
		v, _ := args[key].(string)
		return v
	}
	getInt := func(key string, def int) int {
		v, ok := args[key]
		if !ok {
			return def
		}
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
		return def
	}
	getBool := func(key string, def bool) bool {
		v, ok := args[key]
		if !ok {
			return def
		}
		switch b := v.(type) {
		case bool:
			return b
		case string:
			switch strings.ToLower(strings.TrimSpace(b)) {
			case "true", "1", "yes":
				return true
			case "false", "0", "no":
				return false
			}
		}
		return def
	}

	switch name {
	case "workspace_read_file":
		result, err := r.workspaceReadFile(getString("path"), getInt("start_line", 0), getInt("end_line", 0))
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(result), false, nil
	case "workspace_list_files":
		result, err := r.workspaceListFiles(getString("path"), getBool("recursive", true), getInt("limit", 200), getBool("include_hidden", false))
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(result), false, nil
	case "workspace_search":
		result, err := r.workspaceSearch(getString("query"), getString("path"), getInt("limit", 50), getBool("case_sensitive", false))
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(result), false, nil
	case "workspace_apply_patch":
		result, err := r.workspaceApplyPatch(getString("patch"))
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(result), false, nil
	case "git_status":
		result, err := r.gitStatus()
		if err != nil {
			return err.Error(), true, nil
		}
		return result, false, nil
	case "git_diff":
		result, err := r.gitDiff(getString("path"), getString("ref"), getBool("cached", false))
		if err != nil {
			return err.Error(), true, nil
		}
		return result, false, nil
	case "git_checkout":
		result, err := r.gitCheckout(getString("ref"), getBool("create_branch", false))
		if err != nil {
			return err.Error(), true, nil
		}
		return result, false, nil
	default:
		return "", false, nil
	}
}

func (r *Registry) workspaceReadFile(path string, startLine, endLine int) (*workspaceReadResult, error) {
	resolved, rel, err := r.resolveRepoPath(path, false)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, err
	}
	lines := splitLines(string(data))
	if startLine <= 0 {
		startLine = 1
	}
	if endLine <= 0 || endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > endLine {
		return nil, fmt.Errorf("invalid line range")
	}
	return &workspaceReadResult{
		Path:      rel,
		StartLine: startLine,
		EndLine:   endLine,
		Content:   strings.Join(lines[startLine-1:endLine], "\n"),
	}, nil
}

func (r *Registry) workspaceListFiles(path string, recursive bool, limit int, includeHidden bool) (*workspaceListResult, error) {
	if limit <= 0 {
		limit = 200
	}
	resolved, rel, err := r.resolveRepoPath(path, true)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path must be a directory")
	}

	paths := make([]string, 0, min(limit, 32))
	err = filepath.WalkDir(resolved, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == resolved {
			return nil
		}
		name := entry.Name()
		if !includeHidden && strings.HasPrefix(name, ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if name == ".git" && entry.IsDir() {
			return filepath.SkipDir
		}
		if !recursive {
			parent := filepath.Dir(current)
			if parent != resolved {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if entry.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(r.root, current)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relPath))
		if len(paths) >= limit {
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return nil, err
	}
	sort.Strings(paths)
	return &workspaceListResult{BasePath: rel, Paths: paths}, nil
}

func (r *Registry) workspaceSearch(query, path string, limit int, caseSensitive bool) (*workspaceSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 {
		limit = 50
	}
	resolved, rel, err := r.resolveRepoPath(path, true)
	if err != nil {
		return nil, err
	}
	matches := make([]workspaceSearchMatch, 0, min(limit, 8))
	needle := query
	if !caseSensitive {
		needle = strings.ToLower(query)
	}

	err = filepath.WalkDir(resolved, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		data, err := os.ReadFile(current)
		if err != nil {
			return nil
		}
		if bytes.IndexByte(data, 0) >= 0 {
			return nil
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()
			haystack := line
			if !caseSensitive {
				haystack = strings.ToLower(line)
			}
			if !strings.Contains(haystack, needle) {
				continue
			}
			relPath, err := filepath.Rel(r.root, current)
			if err != nil {
				return err
			}
			matches = append(matches, workspaceSearchMatch{
				Path:       filepath.ToSlash(relPath),
				LineNumber: lineNumber,
				Line:       line,
			})
			if len(matches) >= limit {
				return io.EOF
			}
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return nil, err
	}
	return &workspaceSearchResult{
		Query:    query,
		BasePath: rel,
		Matches:  matches,
	}, nil
}

func (r *Registry) workspaceApplyPatch(patch string) (*applyPatchResult, error) {
	if strings.TrimSpace(patch) == "" {
		return nil, fmt.Errorf("patch is required")
	}
	if !strings.HasSuffix(patch, "\n") {
		patch += "\n"
	}
	if _, err := r.runGit(patch, "apply", "--whitespace=nowarn", "--recount", "--unidiff-zero", "-"); err != nil {
		return nil, err
	}
	status, err := r.gitStatus()
	if err != nil {
		return nil, err
	}
	return &applyPatchResult{
		Applied: true,
		Status:  status,
	}, nil
}

func (r *Registry) gitStatus() (string, error) {
	return r.runGit("", "status", "--short", "--branch")
}

func (r *Registry) gitDiff(path, ref string, cached bool) (string, error) {
	args := []string{"diff"}
	if cached {
		args = append(args, "--cached")
	}
	if ref != "" {
		args = append(args, ref)
	}
	if path != "" {
		_, rel, err := r.resolveRepoPath(path, false)
		if err != nil {
			return "", err
		}
		args = append(args, "--", rel)
	}
	return r.runGit("", args...)
}

func (r *Registry) gitCheckout(ref string, createBranch bool) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("ref is required")
	}
	args := []string{"checkout"}
	if createBranch {
		args = append(args, "-b")
	}
	args = append(args, ref)
	if _, err := r.runGit("", args...); err != nil {
		return "", err
	}
	return r.gitStatus()
}

func (r *Registry) runGit(stdin string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", r.root}, args...)...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *Registry) resolveRepoPath(path string, allowRoot bool) (string, string, error) {
	if strings.TrimSpace(path) == "" {
		if allowRoot {
			return r.root, ".", nil
		}
		return "", "", fmt.Errorf("path is required")
	}
	clean := filepath.Clean(path)
	candidate := clean
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(r.root, clean)
	}
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", "", err
	}
	rel, err := filepath.Rel(r.root, absCandidate)
	if err != nil {
		return "", "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", "", fmt.Errorf("path %q escapes repo root", path)
	}
	if rel == "." && !allowRoot {
		return "", "", fmt.Errorf("path must point to a file or subdirectory")
	}
	return absCandidate, filepath.ToSlash(rel), nil
}

func splitLines(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
