<p align="center">
  <img src="frontend/src/assets/images/logo-universal.png" alt="smb-tools logo" width="120" />
</p>

<h1 align="center">smb-tools</h1>

<p align="center">
  A cross-platform desktop app for Super Mega Baseball 4 franchise management and statistics.
</p>

<p align="center">
  <a href="https://github.com/tbrittain/smb-tools/releases">Download latest release</a>
  &nbsp;·&nbsp;
  <a href="https://tbrittain.github.io/smb-tools/">User documentation</a>
</p>

---

<!-- screenshot goes here -->

## What it does

smb-tools reads your SMB4 save game and gives you a Baseball Reference–style franchise history: season-by-season standings, batting and pitching leaderboards, full player career pages, awards tracking, and a Hall of Fame. One button click imports an entire season — no CSV exports, no multi-step wizards.

> [!IMPORTANT]
> **SMB4 only.** This app does not support SMB3 at this time.

## Download

Grab the latest installer for your platform from the [Releases page](https://github.com/tbrittain/smb-tools/releases).

For usage instructions see the [user documentation](https://tbrittain.github.io/smb-tools/).

## Contributing

smb-tools is a [Wails v2](https://wails.io/) desktop app: a Go backend exposed to a Vue 3 frontend via generated TypeScript bindings, packaged into a native binary.

Layer-specific setup and tooling are documented in their own READMEs:

- [Backend (Go)](internal/README.md)
- [Frontend (Vue 3)](frontend/README.md)

### Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| [Go](https://go.dev/dl/) | 1.26 | Backend language |
| [Node.js](https://nodejs.org/) | 26 | Frontend toolchain |
| [Wails CLI](https://wails.io/docs/gettingstarted/installation) | v2.12.x | Desktop app framework |

### Running in development

```sh
# Run the full app in dev mode with hot reload
wails dev
```

> **Linux note:** On newer distros (e.g. Ubuntu 24.04+), pass the `webkit2_41` build tag if you hit a `webkit2gtk-4.0` pkg-config error:
> ```sh
> wails dev -tags webkit2_41
> ```

### Further reading

- [`docs/`](docs/) for game domain knowledge, other misc internal 
- [`CLAUDE.md`](CLAUDE.md) for coding standards for AI-assisted development
