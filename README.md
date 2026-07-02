# leetcli

A polished terminal dashboard for your LeetCode profile.

## Features

- **Clean TUI interface** with multiple themes
- **Difficulty distribution** — Easy/Medium/Hard bars with percentages
- **Language statistics** — Top languages by problems solved
- **Skills/Tags breakdown** — All problem tags with solve counts
- **Streak tracking** — Current daily streak
- **5 built-in themes**: Default, Dracula, Nord, Tokyo Night, Catppuccin
- **Keyboard navigation** — Tab/arrows to switch panels, `t` to cycle themes
- **Async loading** with spinner
- **Refresh with `r`**
- **YAML config** with theme persistence
- **Production-ready architecture** (clean separation: api/service/ui/config)

## Install

```bash
git clone https://github.com/yourname/leetcli
cd leetcli
go build -o leetcli
```

Create configuration file at `~/.leetcli.yaml`:

```yaml
username: your_leetcode_username
refresh_interval: 300
theme: dracula  # default, dracula, nord, tokyo-night, catppuccin
```

## Controls

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `r` | Refresh profile |
| `t` | Cycle theme |
| `Tab` / `→` | Next panel |
| `Shift+Tab` / `←` | Previous panel |
| `h` / `?` | Toggle help |
| `b` | Problem Browser |
| `Enter` | Open detail view / open problem in browser |
| `Esc` | Close detail view |
| `j` / `k` | Move in Problem Browser |
| `[` / `]` | Prev / next page in Problem Browser |

## Screenshots

### Default Theme
```
LeetCode Dashboard • username
theme: default
Rank: 12345   Reputation: 678   Streak: 42 days

Difficulty              Languages            Skills
Easy     ██████████        60%   120   Go        ████████████  100%   Array        ████████████████  100%
Medium   ██████            30%   60    Python    ████████      65%    Dynamic Prog ████████████      75%
Hard     ████              10%   20    JavaScript ██████       45%    String       ██████████        60%

[r] refresh  [t] theme  [tab/←→] panel  [h] help  [q] quit
```

### Dracula Theme
```
LeetCode Dashboard • username
theme: dracula
Rank: 12345   Reputation: 678   Streak: 42 days

Difficulty              Languages            Skills
Easy     ██████████        60%   120   Go        ████████████  100%   Array        ████████████████  100%
Medium   ██████            30%   60    Python    ████████      65%    Dynamic Prog ████████████      75%
Hard     ████              10%   20    JavaScript ██████       45%    String       ██████████        60%

[r] refresh  [t] theme  [tab/←→] panel  [h] help  [q] quit
```

## Architecture

```
leetcli/
├── main.go                      # Entry point
├── internal/
│   ├── api/client.go            # GraphQL HTTP client
│   ├── config/config.go         # YAML config loader
│   ├── domain/models.go         # Domain models
│   ├── service/profile_service.go  # Business logic
│   └── ui/
│       ├── model.go             # Bubble Tea model + update
│       ├── view.go              # Lipgloss rendering
│       └── theme.go             # Theme definitions
```

## Roadmap

- [x] Problem Browser (list, pagination, open in browser)
- [ ] Contest History with rating graph
- [ ] Daily Challenge tracker
- [ ] Export PNG/JSON
- [ ] GitHub Action for README stats
- [ ] Local problem cache + test runner

## License

MIT