## yt login

Authenticate and save YouTrack credentials

### Synopsis

Authenticate against a YouTrack instance and save the credentials to
~/.config/yt/config.yaml.

Prompts for the base URL and a permanent token unless --url and --token are
given. The token is validated by fetching the current user before anything is
written, so a bad token fails fast and never reaches the config file.

Create a permanent token in YouTrack under:
  Profile -> Account Security -> New token...

The token input is hidden when reading from a terminal. Environment variables
(YOUTRACK_URL, YOUTRACK_TOKEN) still take precedence over the saved config at
runtime.

```
yt login [flags]
```

### Examples

```
  # interactive: prompts for URL and token
  yt login

  # non-interactive
  yt login --url https://youtrack.example.com --token perm:abc123...
```

### Options

```
  -h, --help           help for login
      --token string   permanent token (prompted if omitted)
      --url string     YouTrack base URL (prompted if omitted)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI

