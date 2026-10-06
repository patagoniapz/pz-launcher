// Package ui muestra el progreso al usuario. Intenta abrir una ventana nativa
// de progreso (zenity); si no es posible (p. ej. Linux sin zenity instalado),
// cae de forma transparente a registrar el progreso por consola.
package ui

import (
	"errors"
	"log"
	"strings"

	"github.com/ncruces/zenity"
)

// appTitle es el título de las ventanas. SetVersion le añade la versión.
var appTitle = "PZ Launcher"

// SetVersion añade la versión del launcher al título de las ventanas
// (p. ej. "PZ Launcher 0.0.2"). Llamar antes de abrir cualquier diálogo.
func SetVersion(v string) {
	if v != "" {
		appTitle = "PZ Launcher " + v
	}
}

// Action es la elección del usuario en el diálogo posterior a la actualización.
type Action int

const (
	// ActionQuit: salir sin arrancar el juego.
	ActionQuit Action = iota
	// ActionPlay: arrancar el juego.
	ActionPlay
	// ActionCleanLogs: limpiar los logs (sin arrancar el juego).
	ActionCleanLogs
	// ActionSettings: abrir los ajustes de rendimiento (sin arrancar el juego).
	ActionSettings
	// ActionPatchNotes: mostrar las novedades/parches (sin arrancar el juego).
	ActionPatchNotes
	// ActionOpenLogs: abrir la carpeta de datos de Zomboid (logs) en el
	// explorador de archivos (sin arrancar el juego).
	ActionOpenLogs
)

// Etiquetas del menú principal (también usadas para mapear la elección).
const (
	menuPlay     = "Jugar"
	menuSettings = "Ajustes de rendimiento"
	menuNotes    = "Novedades (parches)"
	menuClean    = "Limpiar logs"
	menuOpenLogs = "Abrir carpeta de logs"
	menuQuit     = "Salir"
)

// menuItem es una opción del menú principal: su etiqueta visible y la acción que
// desencadena.
type menuItem struct {
	label  string
	action Action
}

// AskAction muestra el menú principal (todos los botones en una sola pantalla) y
// devuelve la elección. La opción "Novedades" solo aparece si showNotes es true
// (hay parches que mostrar). instruction es el título grande y content el texto
// explicativo debajo. Si no hay entorno gráfico, arranca el juego (ActionPlay).
func AskAction(instruction, content string, showNotes bool) Action {
	items := []menuItem{
		{menuPlay, ActionPlay},
		{menuSettings, ActionSettings},
	}
	if showNotes {
		items = append(items, menuItem{menuNotes, ActionPatchNotes})
	}
	items = append(items,
		menuItem{menuClean, ActionCleanLogs},
		menuItem{menuOpenLogs, ActionOpenLogs},
		menuItem{menuQuit, ActionQuit},
	)
	return showMenu(instruction, content, items, ActionPlay)
}

// zenityListMenu presenta el menú como una lista nativa de zenity. Es el menú en
// Linux/macOS y el plan B en Windows si el diálogo de botones no está disponible.
// Al cancelar devuelve ActionQuit; sin entorno gráfico, ActionPlay.
func zenityListMenu(instruction, content string, items []menuItem, def Action) Action {
	labels := make([]string, len(items))
	defLabel := items[0].label
	for i, it := range items {
		labels[i] = it.label
		if it.action == def {
			defLabel = it.label
		}
	}
	text := instruction
	if content != "" {
		text += "\n" + content
	}

	choice, err := zenity.List(text, labels,
		zenity.Title(appTitle),
		zenity.DefaultItems(defLabel),
		zenity.DisallowEmpty(),
	)
	if err != nil {
		if errors.Is(err, zenity.ErrCanceled) {
			return ActionQuit
		}
		// Sin GUI disponible (p. ej. consola): comportamiento por defecto.
		return ActionPlay
	}
	for _, it := range items {
		if it.label == choice {
			return it.action
		}
	}
	return ActionPlay
}

// AskText muestra un cuadro de entrada de texto con un mensaje explicativo y un
// valor por defecto. Devuelve el texto introducido y true, o ("", false) si el
// usuario canceló o no hay entorno gráfico.
//
// OJO: el campo de entrada (zenity.Entry) tiene la etiqueta en UNA sola línea de
// ancho fijo; un prompt largo se recorta. Para texto explicativo usa AskNumber,
// que muestra la explicación en un cuadro aparte (que sí hace varias líneas).
func AskText(prompt, def string) (string, bool) {
	s, err := zenity.Entry(prompt, zenity.Title(appTitle), zenity.EntryText(def))
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(s), true
}

// AskNumber pide un valor precedido de una explicación. La explicación (help,
// que puede ocupar varias líneas) se muestra en un cuadro de información aparte
// porque el campo de entrada recorta los textos largos; después se pide el valor
// con una etiqueta corta (label) que SÍ cabe. Devuelve el texto y true, o
// ("", false) si el usuario canceló el campo de entrada.
func AskNumber(help, label, def string) (string, bool) {
	_ = zenity.Info(help, zenity.Title(appTitle), zenity.InfoIcon, zenity.OKLabel("Continuar"))
	return AskText(label, def)
}

// AskYesNo muestra una pregunta con botones Sí/No y devuelve la elección del
// usuario. def es el valor por defecto que se devuelve si no hay entorno gráfico
// o si el usuario cierra el diálogo sin elegir (p. ej. ESC). okLabel/noLabel son
// los textos de los botones (p. ej. "Activado"/"Desactivado").
func AskYesNo(msg, okLabel, noLabel string, def bool) bool {
	err := zenity.Question(msg,
		zenity.Title(appTitle),
		zenity.QuestionIcon,
		zenity.OKLabel(okLabel),
		zenity.CancelLabel(noLabel),
	)
	if err == nil {
		return true // pulsó el botón OK (okLabel)
	}
	if errors.Is(err, zenity.ErrCanceled) {
		return false // pulsó el botón Cancel (noLabel)
	}
	return def // sin GUI u otro error: valor por defecto
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
