// Command pzlauncher-gui es la cara GRÁFICA (Wails) del launcher de Project
// Zomboid para Windows. Reutiliza toda la lógica del launcher de consola
// (paquetes internal/*): descarga del manifiesto, autoactualización, sync de
// ficheros por hash, ajustes de rendimiento y lanzamiento vía Steam. Lo único
// que cambia es la capa de presentación: en vez de los diálogos de zenity,
// abre una ventana propia (WebView2) con progreso y un menú con botones.
//
// El launcher de consola (cmd/launcher) sigue intacto como camino para
// Linux/macOS y como alternativa sin entorno gráfico.
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// Valores inyectados en compilación con -ldflags, igual que en cmd/launcher:
//
//	wails build -ldflags "-X main.version=1.0.0 -X main.manifestURL=https://..."
var (
	version     = "0.0.0-dev"
	manifestURL = "https://github.com/patagoniapz/pz-launcher/releases/latest/download/manifest.json"
)

func main() {
	app := NewApp(version, manifestURL)

	err := wails.Run(&options.App{
		Title:            "PZ Launcher " + version,
		Width:            720,
		Height:           560,
		// Launcher de tamaño fijo: sin redimensionar ni maximizar (el botón de
		// maximizar queda deshabilitado). Se puede mover y minimizar con normalidad.
		DisableResize: true,
		BackgroundColour: &options.RGBA{R: 18, G: 20, B: 24, A: 1},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		Bind:             []interface{}{app},
		Windows: &windows.Options{
			// WebView2 ya viene en Windows 10/11; si faltara, que lo instale.
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
