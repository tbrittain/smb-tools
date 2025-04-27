# README

## About

This is the official Wails Vue-TS template.

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## Requirements

- Go 1.24
- Node 22.15.0
- npm 10.9.2
- [Wails](https://wails.io/docs/gettingstarted/installation) 2.10.1

## Concept

This is a rewrite of the following two projects into a single, unified application:
- [SMB Explorer](https://github.com/tbrittain/SMB3Explorer)
- [SMB Explorer Companion](https://github.com/tbrittain/SmbExplorerCompanion)

The main reasoning behind this is
1. To have a single codebase to maintain
    - Simplify some more complex user flows that require both applications
    - Rewrite the code to support rigorous unit and integration testing
2. Use Go instead of .NET for the backend pieces for increased portability
    - No more runtime requirements for .NET
3. To use Wails for the frontend instead of WPF for a more modern UI experience
    - Increased responsiveness given Webview2 (using Vue.js)
    - Access to mature JavaScript libraries such as D3.js for graphing
    - Better HMR for local development
4. Provide an extensible platform for other purposes related to SMB (roadmap TBD)
    - For example, a file manager for SMB shares with import/export capabilities (this has been a 
      long-standing request)
    - Plugin system (for example, define your own WAR calculation algorithm)

## Live Development

To run in live development mode, run `wails dev` from the root of the repo. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

Generate frontend models with `wails generate models` to get the latest changes from your Go code. These models should
also be generated when you run `wails build` or `wails dev`, but the explicit generate command is nice if you
want to generate models before running the dev server.

## Building

To build a redistributable, production mode package, use `wails build`. (TBD on automation for this)
