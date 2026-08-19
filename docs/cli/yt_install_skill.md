## yt install skill

Install the yt skill for Claude Code and Codex

### Synopsis

Install the YouTrack CLI skill for every supported agent found on this machine.

Supported agents:

  Claude Code   ~/.claude/skills/yt/SKILL.md
  Codex         ~/.codex/skills/yt/SKILL.md

An agent counts as present when its executable is on $PATH or its config
directory holds more than the yt skill itself. Agents that are absent are
skipped, and every agent is reported as installed (nothing was there),
updated (the file existed with older content), unchanged (already current),
or skipped.

The skill frontmatter is adapted per agent: Codex only reads name and
description, so Claude-specific keys are stripped from its copy.

A legacy Claude Code command at ~/.claude/commands/yt.md is removed if present.

```
yt install skill [flags]
```

### Examples

```
  yt install skill
```

### Options

```
  -h, --help   help for skill
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt install](yt_install.md)	 - Install shell completions and agent skills

