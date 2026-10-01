// Package ui muestra el progreso al usuario. Intenta abrir una ventana nativa
// de progreso (zenity); si no es posible (p. ej. Linux sin zenity instalado),
// cae de forma transparente a registrar el progreso por consola.
package ui

import (
	"errors"
	"log"

	"github.com/ncruces/zenity"
)

const appTitle = "PZ Launcher"

// Action es la elección del usuario en el diálogo posterior a la actualización.
type Action int

const (
	// ActionQuit: salir sin arrancar el juego.
	ActionQuit Action = iota
	// ActionPlay: arrancar el juego.
	ActionPlay
	// ActionCleanAndPlay: limpiar los logs y luego arrancar el juego.
	ActionCleanAndPlay
)

// AskAction muestra un diálogo con tres opciones (Jugar / Limpiar logs y jugar /
// Salir) y devuelve la elección. Si no hay entorno gráfico, arranca el juego
// por defecto (ActionPlay).
func AskAction(text string) Action {
	err := zenity.Question(text,
		zenity.Title(appTitle),
		zenity.OKLabel("Jugar"),
		zenity.ExtraButton("Limpiar logs y jugar"),
		zenity.CancelLabel("Salir"),
	)
	switch {
	case err == nil:
		return ActionPlay
	case errors.Is(err, zenity.ErrExtraButton):
		return ActionCleanAndPlay
	case errors.Is(err, zenity.ErrCanceled):
		return ActionQuit
	default:
		// Sin GUI disponible (p. ej. consola): comportamiento por defecto.
		return ActionPlay
	}
}

// PickFolder abre un selector de carpeta. Devuelve la ruta y true si el usuario
// eligió una; false si canceló o no hay entorno gráfico.
func PickFolder(title string) (string, bool) {
	dir, err := zenity.SelectFile(zenity.Directory(), zenity.Title(title))
	if err != nil || dir == "" {
		return "", false
	}
	return dir, true
}

// Info muestra un aviso informativo nativo (y lo registra por consola).
func Info(msg string) {
	log.Println(msg)
	_ = zenity.Info(msg, zenity.Title(appTitle), zenity.InfoIcon)
}

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
