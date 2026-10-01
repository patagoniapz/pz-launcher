// Command launcher es el actualizador/lanzador multiplataforma de Project Zomboid.
//
// Flujo al arrancar:
//  1. Resuelve la carpeta del propio ejecutable (rutas relativas a ella).
//  2. Descarga el manifiesto (manifest.json).
//  3. Si el manifiesto anuncia una versión del launcher más nueva, se
//     autoactualiza y se relanza.
//  4. Sincroniza los ficheros del juego (p. ej. projectzomboid.jar) por hash.
//  5. Lanza el juego.
//
// Muestra el progreso en una ventana nativa (zenity); si no hay entorno gráfico
// disponible, cae a registrar el progreso por consola.
//
// Se deja dentro de la carpeta del juego, p. ej.:
//
//	D:\Program Files (x86)\Steam\steamapps\common\ProjectZomboid\pzlauncher.exe
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"pzlauncher/internal/gamelaunch"
	"pzlauncher/internal/manifest"
	"pzlauncher/internal/paths"
	"pzlauncher/internal/selfupdate"
	"pzlauncher/internal/ui"
	"pzlauncher/internal/updater"
)

// Valores inyectados en tiempo de compilación con -ldflags:
//
//	go build -ldflags "-X main.version=1.0.0 -X main.manifestURL=https://..."
var (
	version     = "0.0.0-dev"
	manifestURL = "https://github.com/patagoniapz/pz-launcher/releases/latest/download/manifest.json"
)

func main() {
	log.SetFlags(0)

	var (
		noLaunch    = flag.Bool("no-launch", false, "actualizar pero no arrancar el juego")
		noSelf      = flag.Bool("no-self-update", false, "no autoactualizar el launcher")
		showVer     = flag.Bool("version", false, "mostrar la versión y salir")
		urlOverride = flag.String("manifest", "", "URL del manifiesto (sobrescribe el valor por defecto)")
	)
	flag.Parse()

	if *showVer {
		fmt.Println("pzlauncher " + version)
		return
	}

	// La URL del manifiesto se puede sobrescribir por flag o por variable de
	// entorno (útil para probar contra un servidor local antes de publicar).
	url := manifestURL
	if env := os.Getenv("PZL_MANIFEST_URL"); env != "" {
		url = env
	}
	if *urlOverride != "" {
		url = *urlOverride
	}

	// Limpia un posible "<exe>.old" dejado por una autoactualización anterior.
	selfupdate.CleanupOld()

	rep := ui.New()
	defer rep.Close()

	baseDir, err := paths.ExecutableDir()
	if err != nil {
		ui.ShowError("No se pudo determinar la carpeta del launcher: " + err.Error())
		os.Exit(1)
	}
	log.Printf("pzlauncher %s | carpeta: %s", version, baseDir)

	rep.Stage("Comprobando actualizaciones…")
	m, err := manifest.Fetch(url)
	if err != nil {
		// Sin conexión no debería impedir jugar: avisamos y arrancamos igual.
		log.Printf("[aviso] no se pudo obtener el manifiesto: %v", err)
		if !*noLaunch {
			rep.Success("Sin conexión. Iniciando el juego…")
			launchBestEffort(baseDir)
		}
		return
	}

	if !*noSelf {
		relaunched, err := selfupdate.MaybeUpdate(version, m, rep)
		if err != nil {
			log.Printf("[aviso] fallo autoactualizando el launcher: %v", err)
		} else if relaunched {
			// El nuevo binario ya está en marcha; este proceso debe salir.
			return
		}
	}

	if err := updater.SyncFiles(baseDir, m.Files, rep); err != nil {
		ui.ShowError("Error actualizando ficheros: " + err.Error())
		os.Exit(1)
	}

	if *noLaunch {
		rep.Success("Actualización completada.")
		return
	}

	rep.Success("¡Listo! Iniciando Project Zomboid…")
	if err := gamelaunch.Launch(baseDir, m); err != nil {
		ui.ShowError(err.Error())
		os.Exit(1)
	}
}

// launchBestEffort arranca el juego sin manifiesto, usando los ejecutables
// típicos de Project Zomboid según el SO. Solo se usa como plan B offline.
func launchBestEffort(baseDir string) {
	fallback := &manifest.Manifest{Launch: map[string]manifest.LaunchSpec{
		"windows": {Exe: "ProjectZomboid64.exe"},
		"linux":   {Exe: "./ProjectZomboid64"},
		"darwin":  {Exe: "./ProjectZomboid64"},
	}}
	if err := gamelaunch.Launch(baseDir, fallback); err != nil {
		log.Printf("[aviso] %v", err)
	}
}
