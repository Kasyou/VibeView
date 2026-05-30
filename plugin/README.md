# VibeView — Claude Code Plugin

Android Studio-style instant UI preview for vibe coding. This plugin gives Claude Code the ability to **see** the rendered output of your frontend project and **self-correct** UI issues in real time.

## Installation

### 1. Install the VibeView binary

```bash
go install github.com/Kasyou/VibeView@latest
```

Or download a prebuilt binary from [GitHub Releases](https://github.com/Kasyou/VibeView/releases) and add it to your PATH.

### 2. Install the plugin

Copy this plugin directory to your Claude Code plugins folder:

```bash
cp -r plugin/ ~/.claude/plugins/cache/local/vibeview/0.1.0/
```

Then enable it in `~/.claude/settings.json`:

```json
{
  "enabledPlugins": {
    "vibeview@local": true
  }
}
```

### 3. Start using

Open any frontend project, start a Claude Code session, and start building UI. Claude will automatically use the preview tools to show you what it builds.

The preview server auto-starts when your session begins. Open http://localhost:51820 to see the live preview.

## What Claude Can Do

| Tool | Capability |
|------|-----------|
| `preview_screenshot` | See the current rendered page as an image |
| `preview_inspect` | Query element position, size, styles, text |
| `preview_console` | Read browser console errors and warnings |
| `preview_diff` | Compare before/after screenshots |
| `preview_reload` | Force refresh the preview |

## Self-Correction

Claude's workflow when building UI:

```
Generate component → preview_screenshot → sees button overflow
→ auto-fixes CSS → preview_screenshot again → verified! → shows result
```

## Requirements

- Go 1.23+ (for `go install`; prebuilt binaries available for all platforms)
- A frontend project (React/Vue/Svelte with Vite, or plain HTML)

## License

MIT
