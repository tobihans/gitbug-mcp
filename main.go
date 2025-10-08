package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GitBugMCPServer provides tools to interact with git-bug via MCP
type GitBugMCPServer struct {
	*mcp.Server
}

// NewGitBugMCPServer creates a new git-bug MCP server
func NewGitBugMCPServer() *GitBugMCPServer {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "gitbug-mcp",
		Version: "1.0.0",
	}, nil)

	gbServer := &GitBugMCPServer{Server: server}

	// Register all tools
	gbServer.registerTools()

	return gbServer
}

// runGitBugCommand executes a git-bug command and returns the output
func runGitBugCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git-bug", args...)

	// Set environment to ensure non-interactive operation
	cmd.Env = append(os.Environ(), "GIT_BUG_NON_INTERACTIVE=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git-bug command failed: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// CreateIssueParams defines parameters for creating a new issue
type CreateIssueParams struct {
	Title   string `json:"title" jsonschema:"required;The title of the issue"`
	Message string `json:"message" jsonschema:"required;The description/message of the issue"`
}

// CreateIssue creates a new bug in git-bug
func (s *GitBugMCPServer) CreateIssue(ctx context.Context, req *mcp.CallToolRequest, params *CreateIssueParams) (*mcp.CallToolResult, any, error) {
	output, err := runGitBugCommand(ctx, "bug", "new", "--title", params.Title, "--message", params.Message, "--non-interactive")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create issue: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	// Extract the bug ID from the output
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var bugID string
	if len(lines) > 0 {
		// The last line usually contains the bug ID or confirmation
		bugID = strings.TrimSpace(lines[len(lines)-1])
	}

	result := map[string]any{
		"success": true,
		"message": "Issue created successfully",
		"title":   params.Title,
		"bug_id":  bugID,
		"output":  strings.TrimSpace(output),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Created issue: %s\n%s", params.Title, bugID)}},
	}, result, nil
}

// ListIssuesParams defines parameters for listing issues
type ListIssuesParams struct {
	Status string `json:"status" jsonschema:"optional;Filter by status (open, closed);default:open"`
	Author string `json:"author" jsonschema:"optional;Filter by author"`
	Label  string `json:"label" jsonschema:"optional;Filter by label"`
	Format string `json:"format" jsonschema:"optional;Output format (default, plain, id, json);default:default"`
	Query  string `json:"query" jsonschema:"optional;Search query string"`
}

// ListIssues lists bugs in git-bug
func (s *GitBugMCPServer) ListIssues(ctx context.Context, req *mcp.CallToolRequest, params *ListIssuesParams) (*mcp.CallToolResult, any, error) {
	args := []string{"bug"}

	// Add filters
	if params.Status != "" {
		args = append(args, "--status", params.Status)
	}
	if params.Author != "" {
		args = append(args, "--author", params.Author)
	}
	if params.Label != "" {
		args = append(args, "--label", params.Label)
	}
	if params.Format != "" {
		args = append(args, "--format", params.Format)
	}
	if params.Query != "" {
		args = append(args, params.Query)
	}

	output, err := runGitBugCommand(ctx, args...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list issues: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	result := map[string]any{
		"success": true,
		"issues":  strings.TrimSpace(output),
		"filters": map[string]string{"status": params.Status, "author": params.Author, "label": params.Label, "format": params.Format, "query": params.Query},
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, result, nil
}

// ShowIssueParams defines parameters for showing an issue
type ShowIssueParams struct {
	BugID  string `json:"bug_id" jsonschema:"required;The ID of the bug to show"`
	Format string `json:"format" jsonschema:"optional;Output format (default, json, org-mode);default:default"`
	Field  string `json:"field" jsonschema:"optional;Specific field to display (author, authorEmail, createTime, lastEdit, humanId, id, labels, shortId, status, title, actors, participants)"`
}

// ShowIssue displays details of a specific bug
func (s *GitBugMCPServer) ShowIssue(ctx context.Context, req *mcp.CallToolRequest, params *ShowIssueParams) (*mcp.CallToolResult, any, error) {
	args := []string{"bug", "show", params.BugID}

	if params.Format != "" {
		args = append(args, "--format", params.Format)
	}
	if params.Field != "" {
		args = append(args, "--field", params.Field)
	}

	output, err := runGitBugCommand(ctx, args...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to show issue %s: %v", params.BugID, err)}},
			IsError: true,
		}, nil, nil
	}

	result := map[string]any{
		"success": true,
		"bug_id":  params.BugID,
		"details": strings.TrimSpace(output),
		"format":  params.Format,
		"field":   params.Field,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, result, nil
}

// AddCommentParams defines parameters for adding a comment
type AddCommentParams struct {
	BugID   string `json:"bug_id" jsonschema:"required;The ID of the bug to comment on"`
	Message string `json:"message" jsonschema:"required;The comment message"`
}

// AddComment adds a comment to a bug
func (s *GitBugMCPServer) AddComment(ctx context.Context, req *mcp.CallToolRequest, params *AddCommentParams) (*mcp.CallToolResult, any, error) {
	output, err := runGitBugCommand(ctx, "bug", "comment", "new", params.BugID, "--message", params.Message, "--non-interactive")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to add comment to bug %s: %v", params.BugID, err)}},
			IsError: true,
		}, nil, nil
	}

	result := map[string]any{
		"success": true,
		"bug_id":  params.BugID,
		"message": "Comment added successfully",
		"output":  strings.TrimSpace(output),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Added comment to bug %s", params.BugID)}},
	}, result, nil
}

// UpdateIssueStatusParams defines parameters for updating issue status
type UpdateIssueStatusParams struct {
	BugID  string `json:"bug_id" jsonschema:"required;The ID of the bug to update"`
	Status string `json:"status" jsonschema:"required;The new status (open, closed)"`
}

// UpdateIssueStatus updates the status of a bug
func (s *GitBugMCPServer) UpdateIssueStatus(ctx context.Context, req *mcp.CallToolRequest, params *UpdateIssueStatusParams) (*mcp.CallToolResult, any, error) {
	var statusCmd string
	switch params.Status {
	case "open":
		statusCmd = "status open"
	case "closed":
		statusCmd = "status close"
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid status: %s. Must be 'open' or 'closed'", params.Status)}},
			IsError: true,
		}, nil, nil
	}

	args := []string{"bug", statusCmd, params.BugID}
	output, err := runGitBugCommand(ctx, args...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update status of bug %s: %v", params.BugID, err)}},
			IsError: true,
		}, nil, nil
	}

	result := map[string]any{
		"success":    true,
		"bug_id":     params.BugID,
		"new_status": params.Status,
		"message":    fmt.Sprintf("Bug status updated to %s", params.Status),
		"output":     strings.TrimSpace(output),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Updated bug %s status to %s", params.BugID, params.Status)}},
	}, result, nil
}

// UpdateIssueTitleParams defines parameters for updating issue title
type UpdateIssueTitleParams struct {
	BugID string `json:"bug_id" jsonschema:"required;The ID of the bug to update"`
	Title string `json:"title" jsonschema:"required;The new title for the bug"`
}

// UpdateIssueTitle updates the title of a bug
func (s *GitBugMCPServer) UpdateIssueTitle(ctx context.Context, req *mcp.CallToolRequest, params *UpdateIssueTitleParams) (*mcp.CallToolResult, any, error) {
	output, err := runGitBugCommand(ctx, "bug", "title", "edit", params.BugID, "--title", params.Title, "--non-interactive")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update title of bug %s: %v", params.BugID, err)}},
			IsError: true,
		}, nil, nil
	}

	result := map[string]any{
		"success":   true,
		"bug_id":    params.BugID,
		"new_title": params.Title,
		"message":   "Bug title updated successfully",
		"output":    strings.TrimSpace(output),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Updated bug %s title to: %s", params.BugID, params.Title)}},
	}, result, nil
}

// DeleteIssueParams defines parameters for deleting an issue
type DeleteIssueParams struct {
	BugID string `json:"bug_id" jsonschema:"required;The ID of the bug to delete"`
}

// HelloGitBugParams defines parameters for the hello git bug tool (no parameters needed)
type HelloGitBugParams struct {
}

// DeleteIssue removes a bug from git-bug
func (s *GitBugMCPServer) DeleteIssue(ctx context.Context, req *mcp.CallToolRequest, params *DeleteIssueParams) (*mcp.CallToolResult, any, error) {
	output, err := runGitBugCommand(ctx, "bug", "rm", params.BugID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete bug %s: %v", params.BugID, err)}},
			IsError: true,
		}, nil, nil
	}

	result := map[string]any{
		"success": true,
		"bug_id":  params.BugID,
		"message": "Bug deleted successfully",
		"output":  strings.TrimSpace(output),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Deleted bug %s", params.BugID)}},
	}, result, nil
}

// HelloGitBug returns a greeting message
func (s *GitBugMCPServer) HelloGitBug(ctx context.Context, req *mcp.CallToolRequest, params *HelloGitBugParams) (*mcp.CallToolResult, any, error) {
	message := "Hello Git Bug"
	
	result := map[string]any{
		"success": true,
		"message": message,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: message}},
	}, result, nil
}

// registerTools registers all git-bug tools with the MCP server
func (s *GitBugMCPServer) registerTools() {
	// Create Issue Tool
	mcp.AddTool(s.Server, &mcp.Tool{
		Name:        "create_issue",
		Description: "Create a new bug in git-bug",
	}, s.CreateIssue)

	// List Issues Tool
	mcp.AddTool(s.Server, &mcp.Tool{
		Name:        "list_issues",
		Description: "List bugs in git-bug with optional filters",
	}, s.ListIssues)

	// Show Issue Tool
	mcp.AddTool(s.Server, &mcp.Tool{
		Name:        "show_issue",
		Description: "Display details of a specific bug",
	}, s.ShowIssue)

	// Add Comment Tool
	mcp.AddTool(s.Server, &mcp.Tool{
		Name:        "add_comment",
		Description: "Add a comment to a bug",
	}, s.AddComment)

	// Update Issue Status Tool
	mcp.AddTool(s.Server, &mcp.Tool{
		Name:        "update_issue_status",
		Description: "Update the status of a bug (open/closed)",
	}, s.UpdateIssueStatus)

	// Update Issue Title Tool
	mcp.AddTool(s.Server, &mcp.Tool{
		Name:        "update_issue_title",
		Description: "Update the title of a bug",
	}, s.UpdateIssueTitle)

	// Delete Issue Tool
	mcp.AddTool(s.Server, &mcp.Tool{
		Name:        "delete_issue",
		Description: "Remove a bug from git-bug",
	}, s.DeleteIssue)
}

func main() {
	server := NewGitBugMCPServer()

	log.Printf("Starting git-bug MCP server...")

	// Run the server with stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
