# GoBuilder

> Cross-compile your Go project for every platform in one command.

GoBuilder is a CLI tool that cross-compiles your Go project for multiple platforms and architectures in parallel. Instead of manually setting `GOOS` and `GOARCH` for every target, GoBuilder discovers your project's entry point, resolves all available Go toolchain targets, and builds them concurrently — then prints a live progress counter and a formatted summary table with file sizes.

---

## Features

- **Parallel builds** — runs up to `N` builds at once (defaults to your CPU core count)
- **Auto-detection** — finds your `main()` entry point automatically, no config needed
- **Live progress** — animated spinner with a real-time `✓ / ✗` build counter
- **Summary table** — sorted output table with target, status, filename, and file size
- **Compression** — optionally compresses each binary (`.zip` for Windows, `.tar.gz` for others)
- **Dry run** — preview exactly which targets would be built before committing
- **First-class filter** — limit builds to Go's officially supported tier-1 platforms
- **Friendly errors** — validates your project, entry file, and Go installation before building

---

## Installation

**Via `go install`:**
```bash
go install github.com/Moukhtar-youssef/GoBuilder@latest
```

**From a release binary:**

Download the archive for your platform from the [Releases](https://github.com/Moukhtar-youssef/GoBuilder/releases) page, extract it, and move the binary to a directory in your `PATH`.

```bash
# Example for Linux amd64
tar -xzf GoBuilder_linux_amd64.tar.gz
mv GoBuilder_linux_amd64 /usr/local/bin/GoBuilder
```

---

## Requirements

- Go 1.23 or later (uses `strings.SplitSeq`)
- Go toolchain available in your `PATH`

---

## Usage

```
GoBuilder [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--project` | `-p` | `.` | Path to the Go project directory |
| `--output` | `-o` | `./dist` | Directory to write build outputs |
| `--targets` | `-t` | all | Comma-separated targets e.g. `linux/amd64,darwin/arm64` |
| `--name` | `-n` | folder name | Output binary name |
| `--entry` | `-e` | auto-detected | Path to a specific entry file |
| `--concurrency` | `-c` | CPU cores | Max parallel builds |
| `--compress` | `-z` | false | Compress each binary after building |
| `--first-class-only` | | false | Only build for first-class Go platforms |
| `--dry-run` | | false | Preview targets without building |
| `--version` | `-v` | | Print version |

---

## Examples

```bash
# Build for every available platform
GoBuilder -p ./myproject

# Build for specific targets only
GoBuilder -p ./myproject -t linux/amd64,darwin/arm64,windows/amd64

# Build first-class platforms only with a custom name
GoBuilder -p ./myproject --first-class-only -n mybinary

# Build and compress all outputs into archives
GoBuilder -p ./myproject --compress -o ./releases

# Preview what would be built without running any builds
GoBuilder -p ./myproject --first-class-only --dry-run

# Limit concurrency and point at a specific entry file
GoBuilder -p ./myproject -e cmd/server/main.go -c 4
```

---

## Output

```
✓ Built 6 / 6 targets

TARGET            STATUS   OUTPUT                            SIZE
------            ------   ------                            ----
darwin/amd64      ✓        mybinary_darwin_amd64.tar.gz      2.8 MB
darwin/arm64      ✓        mybinary_darwin_arm64.tar.gz      2.7 MB
linux/amd64       ✓        mybinary_linux_amd64.tar.gz       2.8 MB
linux/arm64       ✓        mybinary_linux_arm64.tar.gz       2.6 MB
windows/amd64     ✓        mybinary_windows_amd64.zip        2.9 MB
windows/arm64     ✓        mybinary_windows_arm64.zip        2.8 MB
```

---

## How it works

1. Validates the project directory and locates the `main()` entry point via AST parsing
2. Fetches all available platforms from `go tool dist list`
3. Filters targets based on your flags
4. Spawns parallel build goroutines limited by a semaphore channel
5. Collects results into a `buildResult` slice — no output prints mid-build
6. Optionally compresses each binary after all builds complete
7. Prints the summary table

---

## Building from source

```bash
git clone https://github.com/Moukhtar-youssef/GoBuilder.git
cd GoBuilder
go build -ldflags "-X github.com/Moukhtar-youssef/GoBuilder/cmd.version=dev" -o GoBuilder .
```

---

## License

MIT
