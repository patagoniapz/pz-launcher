// Command launcher es el actualizador/lanzador multiplataforma de Project Zomboid.
//
// Flujo al arrancar:
//  1. Descarga el manifiesto (manifest.json).
//  2. Si el manifiesto anuncia una versión del launcher más nueva, se
//     autoactualiza y se relanza.
//  3. Localiza la carpeta de INSTALACIÓN del juego (autodetección vía Steam; si
//     no, pregunta y la recuerda) y sincroniza sus ficheros (p. ej.
//     projectzomboid.jar) por hash, mostrando progreso y tiempo estimado.
//  4. Pregunta al usuario qué hacer: Jugar / Limpiar logs / Salir.
//  5. Lanza el juego vía Steam (arranca Steam si está cerrado).
//
// El ejecutable NO necesita estar dentro de la carpeta del juego: detecta dónde
// está instalado. Muestra el progreso en una ventana nativa (zenity); si no hay
// entorno gráfico, cae a registrar el progreso por consola.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"pzlauncher/internal/config"
	"pzlauncher/internal/gamelaunch"
	"pzlauncher/internal/manifest"
	"pzlauncher/internal/paths"
	"pzlauncher/internal/selfupdate"
	"pzlauncher/internal/steam"
	"pzlauncher/internal/ui"
	"pzlauncher/internal/updater"
	"pzlauncher/internal/zomboid"
)

const pzAppID = 108600 // Project Zomboid en Steam

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
		detect      = flag.Bool("detect", false, "mostrar la carpeta del juego detectada y salir")
		urlOverride = flag.String("manifest", "", "URL del manifiesto (sobrescribe el valor por defecto)")
	)
	flag.Parse()

	if *showVer {
		fmt.Println("pzlauncher " + version)
		return
	}

	if *detect {
		if dir, ok := steam.FindGameDir(pzAppID); ok {
			fmt.Println(dir)
		} else {
			fmt.Println("No se detectó la carpeta de Project Zomboid.")
		}
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

	exeDir, err := paths.ExecutableDir()
	if err != nil {
		ui.ShowError("No se pudo determinar la carpeta del launcher: " + err.Error())
		os.Exit(1)
	}
	cfg := config.Load(exeDir)
	log.Printf("pzlauncher %s | carpeta del exe: %s", version, exeDir)

	rep.Stage("Comprobando actualizaciones…")
	m, err := manifest.Fetch(url)
	offline := err != nil
	var gameDir string
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

		appID := m.SteamAppID
		if appID == 0 {
			appID = pzAppID
		}
		gameDir, err = resolveGameDir(exeDir, &cfg, appID)
		if err != nil {
			ui.ShowError(err.Error())
			os.Exit(1)
		}
		log.Printf("carpeta del juego: %s", gameDir)

		if err := updater.SyncFiles(gameDir, m.Files, rep); err != nil {
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
			launchGame(gameDir, m)
			return
		case ui.ActionCleanLogs:
			// Limpia y vuelve a mostrar el diálogo (no arranca el juego).
			cleanLogs(exeDir, &cfg)
		}
	}
}

// resolveGameDir localiza la carpeta de instalación del juego con esta prioridad:
//  1. Variable de entorno PZL_GAME_DIR.
//  2. Valor guardado en la config (pzlauncher.json).
//  3. El propio exe ya está dentro de la carpeta del juego.
//  4. Autodetección vía Steam.
//  5. Preguntar al usuario (y recordar la elección).
func resolveGameDir(exeDir string, cfg *config.Config, appID int) (string, error) {
	if env := os.Getenv("PZL_GAME_DIR"); env != "" {
		return env, nil
	}
	if cfg.GameDir != "" && dirExists(cfg.GameDir) {
		return cfg.GameDir, nil
	}
	if hasGameExe(exeDir) {
		return exeDir, nil
	}
	if dir, ok := steam.FindGameDir(appID); ok {
		return dir, nil
	}
	if picked, ok := ui.PickFolder("Selecciona la carpeta de instalación de Project Zomboid"); ok {
		cfg.GameDir = picked
		_ = config.Save(exeDir, *cfg)
		return picked, nil
	}
	return "", errors.New("No se encontró la carpeta de instalación de Project Zomboid.")
}

// hasGameExe indica si dir parece la carpeta del juego (contiene su ejecutable).
func hasGameExe(dir string) bool {
	for _, name := range []string{"ProjectZomboid64.exe", "ProjectZomboid64"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

func dirExists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// cleanLogs borra los logs del cliente de la carpeta de datos de Zomboid.
// Si no se encuentra, pide al usuario que la seleccione y recuerda la elección.
func cleanLogs(exeDir string, cfg *config.Config) {
	dir := zomboid.Resolve(cfg.ZomboidDir)
	if !zomboid.Exists(dir) {
		picked, ok := ui.PickFolder("Selecciona tu carpeta Zomboid (donde están los logs)")
		if !ok {
			ui.ShowError("No se encontró la carpeta de Zomboid; no se limpiaron logs.")
			return
		}
		dir = picked
		cfg.ZomboidDir = picked
		_ = config.Save(exeDir, *cfg)
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

func launchGame(gameDir string, m *manifest.Manifest) {
	if err := gamelaunch.Launch(gameDir, m); err != nil {
		ui.ShowError(err.Error())
		os.Exit(1)
	}
}

// offlineManifest da el mínimo para poder lanzar el juego sin conexión.
func offlineManifest() *manifest.Manifest {
	exe := "ProjectZomboid64.exe"
	if runtime.GOOS != "windows" {
		exe = "./ProjectZomboid64"
	}
	return &manifest.Manifest{
		SteamAppID: pzAppID,
		Launch:     map[string]manifest.LaunchSpec{runtime.GOOS: {Exe: exe}},
	}
}
