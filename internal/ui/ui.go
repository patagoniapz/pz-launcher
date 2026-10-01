// Package ui muestra el progreso al usuario. Intenta abrir una ventana nativa
// de progreso (zenity); si no es posible (p. ej. Linux sin zenity instalado),
// cae de forma transparente a registrar el progreso por consola.
package ui

import (
	"log"

	"github.com/ncruces/zenity"
)

const appTitle = "PZ Launcher"

// Reporter recibe las actualizaciones de estado del launcher.
type Reporter interface {
	// Stage fija el texto de la etapa actual (p. ej. "Descargando ...").
	Stage(text string)
	// Progress informa del avance de la descarga en curso (bytes).
	Progress(done, total int64)
	// Success marca la finalización correcta (barra al 100%).
	Success(text string)
	// Close cierra la ventana / libera recursos.
	Close()
}

// New crea el mejor Reporter disponible: ventana de progreso si se puede,
// o consola como alternativa.
func New() Reporter {
	dlg, err := zenity.Progress(
		zenity.Title(appTitle),
		zenity.MaxValue(100),
		zenity.NoCancel(),
	)
	if err != nil {
		return &consoleReporter{}
	}
	_ = dlg.Text("Iniciando…")
	return &guiReporter{dlg: dlg}
}

// ShowError muestra un cuadro de error nativo (y lo registra por consola).
func ShowError(msg string) {
	log.Println("ERROR:", msg)
	_ = zenity.Error(msg, zenity.Title(appTitle), zenity.ErrorIcon)
}

// --- Implementación GUI (zenity) ---

type guiReporter struct {
	dlg zenity.ProgressDialog
}

func (g *guiReporter) Stage(text string) { _ = g.dlg.Text(text) }

func (g *guiReporter) Progress(done, total int64) {
	if total > 0 {
		pct := int(float64(done) / float64(total) * 100)
		if pct > 100 {
			pct = 100
		}
		_ = g.dlg.Value(pct)
	}
}

func (g *guiReporter) Success(text string) {
	_ = g.dlg.Text(text)
	_ = g.dlg.Value(100)
	_ = g.dlg.Complete()
}

func (g *guiReporter) Close() { _ = g.dlg.Close() }

// --- Implementación consola ---

type consoleReporter struct{ lastPct int }

func (c *consoleReporter) Stage(text string) { log.Println(text) }

func (c *consoleReporter) Progress(done, total int64) {
	if total <= 0 {
		return
	}
	pct := int(float64(done) / float64(total) * 100)
	// Evita inundar la consola: solo registra cada salto de 10%.
	if pct >= c.lastPct+10 || pct == 100 {
		c.lastPct = pct
		log.Printf("  %d%%", pct)
	}
}

func (c *consoleReporter) Success(text string) { log.Println(text) }

func (c *consoleReporter) Close() {}
