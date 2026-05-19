package main

import (
	"context"
	"database/sql"
	"log"

	"smb-tools/internal/config"
	"smb-tools/internal/db"
	"smb-tools/internal/store"
)

// App is the Wails application struct. It is intentionally thin: it wires
// dependencies at startup and exposes bindings to the frontend. All business
// logic lives in internal/service and internal/store.
type App struct {
	ctx            context.Context
	dirs           *config.AppDirs
	registryDB     *sql.DB
	franchiseStore *store.FranchiseStore
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dirs, err := config.NewAppDirs()
	if err != nil {
		log.Printf("startup: resolving app directories: %v", err)
		return
	}
	a.dirs = dirs

	registryDB, err := db.OpenRegistry(ctx, dirs.RegistryPath)
	if err != nil {
		log.Printf("startup: opening registry DB: %v", err)
		return
	}
	a.registryDB = registryDB
	a.franchiseStore = store.NewFranchiseStore(registryDB)
}

func (a *App) shutdown(_ context.Context) {
	if a.registryDB != nil {
		if err := a.registryDB.Close(); err != nil {
			log.Printf("shutdown: closing registry DB: %v", err)
		}
	}
}
