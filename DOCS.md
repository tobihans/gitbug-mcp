# Git-Bug MCP Server

A Model Context Protocol (MCP) server that provides tools to interact with [git-bug](https://github.com/git-bug/git-bug), a distributed bug tracker embedded in Git.

## Features

The server provides the following MCP tools:

### Core Issue Management

1. **`create_issue`** - Create a new bug
   - Parameters: `title` (required), `message` (required)
   - Creates a new bug with title and description

2. **`list_issues`** - List bugs with optional filtering
   - Parameters: `status`, `author`, `label`, `format`, `query` (all optional)
   - Supports filtering by status (open/closed), author, labels, and custom queries

3. **`show_issue`** - Display detailed information about a specific bug
   - Parameters: `bug_id` (required), `format`, `field` (optional)
   - Shows complete bug details or specific fields

4. **`delete_issue`** - Remove a bug from the repository
   - Parameters: `bug_id` (required)
   - Permanently removes the bug (note: this is local-only for bridge-imported bugs)

### Comment Management

5. **`add_comment`** - Add a comment to an existing bug
   - Parameters: `bug_id` (required), `message` (required)
   - Adds a new comment to the specified bug

### Status Management

6. **`update_issue_status`** - Change bug status
   - Parameters: `bug_id` (required), `status` (required: "open" or "closed")
   - Opens or closes the specified bug

### Title Management

7. **`update_issue_title`** - Edit bug title
   - Parameters: `bug_id` (required), `title` (required)
   - Updates the title of the specified bug

## Installation

1. Ensure you have `git-bug` installed and available in your PATH
2. Build the MCP server:
   ```bash
   go build -o gitbug-mcp
   ```

## Usage

### As a standalone MCP server

Run the server with stdio transport:
```bash
./gitbug-mcp
```

### With Claude Desktop

Add to your Claude Desktop configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "gitbug-mcp": {
      "command": "/path/to/gitbug-mcp",
      "args": []
    }
  }
}
```

### Example Tool Usage

#### Create a new issue:
```json
{
  "tool": "create_issue",
  "arguments": {
    "title": "Bug in authentication flow",
    "message": "Users cannot login with valid credentials on the production environment."
  }
}
```

#### List all open issues:
```json
{
  "tool": "list_issues",
  "arguments": {
    "status": "open",
    "format": "default"
  }
}
```

#### Show issue details:
```json
{
  "tool": "show_issue",
  "arguments": {
    "bug_id": "abc123def456",
    "format": "default"
  }
}
```

#### Add a comment:
```json
{
  "tool": "add_comment",
  "arguments": {
    "bug_id": "abc123def456",
    "message": "I can reproduce this issue on my machine."
  }
}
```

#### Close an issue:
```json
{
  "tool": "update_issue_status",
  "arguments": {
    "bug_id": "abc123def456",
    "status": "closed"
  }
}
```

## Requirements

- Go 1.25.1 or later
- `git-bug` installed and in PATH
- A git repository with git-bug initialized

## Security Notes

- The server executes `git-bug` commands with `GIT_BUG_NON_INTERACTIVE=1` to prevent interactive prompts
- All commands are executed with context cancellation support
- The server sanitizes inputs but relies on git-bug's own security model

## Error Handling

All tools return structured error responses:
- Success: Returns the tool result with relevant data
- Error: Returns an error message with details about what went wrong

## Development

To modify or extend the server:

1. Add new tool functions following the existing pattern
2. Register them in `registerTools()`
3. Update parameter structs with proper JSON schema tags
4. Test with your MCP client

## License

MIT License - see LICENSE file for details.