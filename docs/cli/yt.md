## yt

YouTrack CLI

### Synopsis

Command-line interface for JetBrains YouTrack.

Fetch issues, list and filter them, and output as human-readable text or JSON.

Configuration is read from environment variables or ~/.config/yt/config.yaml:

  Environment variables:
    YOUTRACK_URL     Base URL of the YouTrack instance
    YOUTRACK_TOKEN   Permanent token for authentication

  Config file (~/.config/yt/config.yaml):
    url: https://youtrack.example.com
    token: perm:abc123...

Environment variables take precedence over the config file.

### Options

```
  -h, --help   help for yt
      --json   output raw JSON
```

### SEE ALSO

* [yt attachment](yt_attachment.md)	 - Manage issue attachments
* [yt board](yt_board.md)	 - Show board issues or list boards
* [yt branch](yt_branch.md)	 - Create git branch from issue
* [yt install](yt_install.md)	 - Install shell completions and Claude Code skill
* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues
* [yt link](yt_link.md)	 - Create links between issues
* [yt links](yt_links.md)	 - List an issue's links
* [yt login](yt_login.md)	 - Authenticate and save YouTrack credentials
* [yt project](yt_project.md)	 - Inspect YouTrack project details
* [yt projects](yt_projects.md)	 - List YouTrack projects
* [yt unlink](yt_unlink.md)	 - Remove a link between issues

