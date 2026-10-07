package main

import (
	"context"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// wailsReporter implementa pzlauncher/internal/ui.Reporter emitiendo el progreso
// como eventos de Wails ("stage", "progress", "success") que el frontend escucha
// para actualizar la barra. Así los paquetes updater/selfupdate no saben nada de
// la GUI: siguen hablando con la misma interfaz Reporter que con zenity.
type wailsReporter struct{ ctx context.Context }

// Stage fija el texto de la etapa actual (p. ej. "Descargando …").
func (r *wailsReporter) Stage(text string) {
	wruntime.EventsEmit(r.ctx, "stage", text)
}

// Progress informa del avance de la descarga en curso (bytes) con el % ya
// calculado para comodidad del frontend.
func (r *wailsReporter) Progress(done, total int64) {
	pct := 0
	if total > 0 {
		pct = int(float64(done) / float64(total) * 100)
		if pct > 100 {
			pct = 100
		}
	}
	wruntime.EventsEmit(r.ctx, "progress", map[string]any{
		"done":  done,
		"total": total,
		"pct":   pct,
	})
}

// Success marca la finalización correcta (barra al 100%).
func (r *wailsReporter) Success(text string) {
	wruntime.EventsEmit(r.ctx, "success", text)
}

// Close no hace nada: la transición al menú la dispara el evento "ready".
func (r *wailsReporter) Close() {}
