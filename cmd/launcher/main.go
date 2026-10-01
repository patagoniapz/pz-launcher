// Command launcher es el actualizador/lanzador multiplataforma de Project Zomboid.
//
// Flujo al arrancar:
//  1. Resuelve la carpeta del propio ejecutable (rutas relativas a ella).
//  2. Descarga el manifiesto (manifest.json).
//  3. Si el manifiesto anuncia una versión del launcher más nueva, se
//     autoactualiza y se relanza.
//  4. Sincroniza los ficheros del juego (p. ej. projectzomboid.jar) por hash,
//     mostrando progreso y tiempo estimado.
//  5. Pregunta al usuario qué hacer: Jugar / Limpiar logs / Salir.
//  6. Lanza el juego vía Steam (arranca Steam si está cerrado).
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

	"pzlauncher/internal/config"
	"pzlauncher/internal/gamelaunch"
	"pzlauncher/internal/manifest"
	"pzlauncher/internal/paths"
	"pzlauncher/internal/selfupdate"
	"pzlauncher/internal/ui"
	"pzlauncher/internal/updater"
	"pzlauncher/internal/zomboid"
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

	ui.SetVersion(version) // la versión aparece en el título de las ventanas
	rep := ui.New()
	defer rep.Close()

	baseDir, err := paths.ExecutableDir()
	if err != nil {
		ui.ShowError("No se pudo determinar la carpeta del launcher: " + err.Error())
		os.Exit(1)
	}
	cfg := config.Load(baseDir)
	log.Printf("pzlauncher %s | carpeta: %s", version, baseDir)

	rep.Stage("Comprobando actualizaciones…")
	m, err := manifest.Fetch(url)
	offline := err != nil
	if offline {
		// Sin conexión no debería impedir jugar: avisamos y seguimos con un
		// manifiesto mínimo para poder lanzar el juego igualmente.
		log.Printf("[aviso] no se pudo obtener el manifiesto: %v", err)
		m = offlineManifest()
	} else {
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
	}

	// Cerramos el diálogo de progreso antes de preguntar la acción.
	rep.Close()

	if *noLaunch {
		return
	}

	prompt := "Project Zomboid está listo."
	if offline {
		prompt = "Sin conexión (no se comprobaron actualizaciones)."
	}
	prompt += "\n¿Qué quieres hacer?"

	for {
		switch ui.AskAction(prompt) {
		case ui.ActionQuit:
			return
		case ui.ActionPlay:
			launchGame(baseDir, m)
			return
		case ui.ActionCleanLogs:
			// Limpia y vuelve a mostrar el diálogo (no arranca el juego).
			cleanLogs(baseDir, &cfg)
		}
	}
}

// cleanLogs borra los logs del cliente de la carpeta de datos de Zomboid.
// Si no se encuentra, pide al usuario que la seleccione y recuerda la elección.
func cleanLogs(baseDir string, cfg *config.Config) {
	dir := zomboid.Resolve(cfg.ZomboidDir)
	if !zomboid.Exists(dir) {
		picked, ok := ui.PickFolder("Selecciona tu carpeta Zomboid (donde están los logs)")
		if !ok {
			ui.ShowError("No se encontró la carpeta de Zomboid; no se limpiaron logs.")
			return
		}
		dir = picked
		cfg.ZomboidDir = picked
		_ = config.Save(baseDir, *cfg)
	}

	removed, err := zomboid.CleanLogs(dir)
	if err != nil {
		ui.ShowError("Error limpiando logs: " + err.Error())
		return
	}
	log.Printf("Logs limpiados: %d elemento(s) en %s", removed, dir)
	if removed == 0 {
		ui.Info("No había logs que limpiar.")
	} else {
		ui.Info(fmt.Sprintf("Logs limpiados: %d elemento(s).", removed))
	}
}

func launchGame(baseDir string, m *manifest.Manifest) {
	if err := gamelaunch.Launch(baseDir, m); err != nil {
		ui.ShowError(err.Error())
		os.Exit(1)
	}
}

// offlineManifest da el mínimo para poder lanzar el juego sin conexión.
func offlineManifest() *manifest.Manifest {
	return &manifest.Manifest{
		SteamAppID: 108600, // Project Zomboid
		Launch: map[string]manifest.LaunchSpec{
			"windows": {Exe: "ProjectZomboid64.exe"},
			"linux":   {Exe: "./ProjectZomboid64"},
			"darwin":  {Exe: "./ProjectZomboid64"},
		},
	}
}
