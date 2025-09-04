# 🎨 phaser-unpacker

A CLI tool for **unpacking Phaser texture atlases** into individual `.png` sprite images.
Given a `.json` atlas definition and its associated image sheet(s), `phaser-unpacker` extracts every frame and saves it as a standalone file.

---

## Features

- 📂 Reads Phaser `.json` texture pack files.
- 🖼️ Supports `.webp`, `.png`, and other raster formats supported by Go’s image decoders.
- ✂️ Cuts out individual textures from the atlas using metadata coordinates.
- 💾 Saves each extracted texture as a **standalone PNG**.
- ⚡ Runs **concurrently** to speed up large sheets (e.g. 1000+ textures).
- 📊 Optional progress bars to track per-sheet and total progress.

---

## Installation

Clone and build the tool with Go:

```bash
git clone https://github.com/evaneliasyoung/phaser-unpacker.git
cd phaser-unpacker
go build -o phaser-unpacker
```

Or install directly:

```bash
go install github.com/evaneliasyoung/phaser-unpacker@latest
```

Or download the correct [release](https://github.com/evaneliasyoung/phaser-unpacker/releases) for your OS.

---

## Usage

Run the tool with a Phaser atlas `.json` file:

```bash
./phaser-unpacker <path-to-pack.json> [flags]
```

### Required Arguments

| Argument | Description                                  | Example               |
| -------- | -------------------------------------------- | --------------------- |
| `<path>` | Path to the **Phaser atlas JSON** definition | `assets/sprites.json` |

### Optional Flags

| Flag                  | Description                          | Default                  |
| --------------------- | ------------------------------------ | ------------------------ |
| `-o, --output <dir>`  | Directory to write unpacked textures | `<packname>`             |
| `-w, --workers <num>` | Number of concurrent workers         | 2×Thread Count, up to 32 |
| `--no-progress`       | Disables progress bars               | disabled if non-TTY      |

---

## Examples

```bash
# Unpack textures to ./sprites_output/*.png
./phaser-unpacker assets/sprites.json -o sprites_output

# Use 8 workers regardless of CPU count
./phaser-unpacker assets/sprites.json --workers 8

# Run with progress bars
./phaser-unpacker assets/sprites.json
```

---

## Dependencies

- [`spf13/cobra`](https://github.com/spf13/cobra) — CLI framework
- [`golang.org/x/image/webp`](https://pkg.go.dev/golang.org/x/image/webp) — WEBP decoder
- [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term) — Determine if TTY
- [`vbauerster/mpb`](https://github.com/vbauerster/mpb) — Progress bars
