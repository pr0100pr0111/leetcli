# leetcli

A polished terminal dashboard for your LeetCode profile, plus a local test runner.

## Features

- **Clean TUI interface** with multiple themes
- **Difficulty distribution** — Easy/Medium/Hard bars with percentages
- **Language statistics** — Top languages by problems solved
- **Skills/Tags breakdown** — All problem tags with solve counts
- **Streak tracking** — Current daily streak
- **5 built-in themes**: Default, Dracula, Nord, Tokyo Night, Catppuccin
- **Keyboard navigation** — Tab/arrows to switch panels, `t` to cycle themes
- **Problem filter** — `/` live search by title, slug or ID in the Problem Browser
- **Daily Challenge** — today's problem on the dashboard, `d` opens a detail view with topics and link
- **Local test runner** — `leetcli init` scaffolds a solution, `leetcli test` runs it against the sample case (Go, Python 3, JavaScript, C++)
- **Async loading** with spinner
- **Refresh with `r`**
- **YAML config** with theme persistence
- **Production-ready architecture** (clean separation: api/service/runner/cli/ui/config)

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

## Local tests

```bash
leetcli init two-sum --lang python3   # creates solution.py + leetcli.json
leetcli init daily                    # same, but for today's Daily Challenge
$EDITOR solution.py
leetcli test                          # runs the sample case, prints PASS/FAIL
```

`init` fetches the problem and writes the official LeetCode stub into the current
directory; `test` fetches the problem again (signature, sample input, expected
output), runs your solution in a sandbox and compares the result.

- Languages: `golang`, `python3`, `javascript`, `cpp` (aliases: `go`, `python`,
  `js`, `c++`) — the respective toolchain must be on `PATH`
- `leetcli init <slug> --force` overwrites existing files
- `leetcli init daily` resolves today's Daily Challenge to its slug first
- `leetcli test --keep` keeps the generated harness for inspection
- Exit codes: `0` passed, `1` failed, `2` usage error
- No config file needed for `init`/`test`; run `leetcli help` for the reference

Current limitations (v1):

- The expected output is taken from the first example in the statement, so it may
  differ from LeetCode's hidden sample test
- Problems with multiple valid answers (e.g. any valid permutation) may report FAIL
- Debug prints are fine, but avoid printing without a trailing newline (it can
  break result parsing)

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
| `d` | Daily Challenge detail view (`enter` open in browser, `esc` back) |
| `Enter` | Open detail view / open problem in browser |
| `Esc` | Close detail view |
| `j` / `k` | Move in Problem Browser |
| `[` / `]` | Prev / next page in Problem Browser |
| `/` | Filter problems (search by title/slug/ID, `enter` apply, `esc` clear) |

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
├── main.go                      # Entry point (TUI or CLI subcommand)
├── internal/
│   ├── api/                     # GraphQL HTTP clients (profile, problem detail)
│   ├── config/config.go         # YAML config loader
│   ├── domain/                  # Domain models
│   ├── service/                 # Business logic + statement parsing
│   ├── runner/                  # Multi-language local test runner
│   ├── cli/                     # `leetcli init` / `leetcli test`
│   └── ui/
│       ├── model.go             # Bubble Tea model + update
│       ├── view.go              # Lipgloss rendering
│       └── theme.go             # Theme definitions
```

## Roadmap

- [x] Problem Browser (list, pagination, open in browser)
- [x] Local test runner (`leetcli init` / `leetcli test`, Go/Python/JS/C++)
- [x] Daily Challenge (`d` detail view, `leetcli init daily`)
- [ ] Contest History with rating graph
- [ ] Export PNG/JSON
- [ ] GitHub Action for README stats
- [ ] Local problem cache + more test languages (Java, Rust)

## License

MIT