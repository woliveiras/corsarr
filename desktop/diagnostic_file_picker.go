package main

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type wailsDiagnosticFilePicker struct{}

func (wailsDiagnosticFilePicker) Choose(ctx context.Context, suggestedName string) (string, error) {
	return wailsruntime.SaveFileDialog(ctx, wailsruntime.SaveDialogOptions{
		Title:           "Salvar diagnóstico do Corsarr",
		DefaultFilename: suggestedName,
		Filters: []wailsruntime.FileFilter{{
			DisplayName: "Diagnóstico JSON (*.json)",
			Pattern:     "*.json",
		}},
	})
}
