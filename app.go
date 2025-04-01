package main

import (
	"context"
	"fmt"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type Person struct {
	Name    string   `json:"name"`
	Age     uint32   `json:"age"`
	Address *Address `json:"address"`
}

type Address struct {
	Street   string `json:"street"`
	Postcode string `json:"postcode"`
}

// Greet returns a greeting for the given name
func (a *App) Greet(p Person) string {
	return fmt.Sprintf("Hello %s (Age: %d)!", p.Name, p.Age)
}
