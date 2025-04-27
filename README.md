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

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

Generate frontend models with `wails generate models` to get the latest changes from your Go code.

## Building

To build a redistributable, production mode package, use `wails build`.
