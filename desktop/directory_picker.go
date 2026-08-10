package main

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type wailsDirectoryPicker struct{}

func (wailsDirectoryPicker) Choose(ctx context.Context) (string, error) {
	return wailsruntime.OpenDirectoryDialog(ctx, wailsruntime.OpenDialogOptions{
		Title:                "Escolha onde o Corsarr guardará seus arquivos",
		CanCreateDirectories: true,
		ResolvesAliases:      true,
	})
}
