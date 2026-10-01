// Package gamelaunch arranca el juego una vez los ficheros están actualizados.
//
// Preferimos lanzarlo a través de Steam (steam://rungameid/<appid>): así, si
// Steam está cerrado, se abre automáticamente y luego arranca el juego — el
// ejecutable directo (ProjectZomboid64.exe) falla si Steam no está en marcha.
// Si el manifiesto no trae SteamAppID, caemos al ejecutable de LaunchSpec.
package gamelaunch

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"

	"pzlauncher/internal/manifest"
)

// Launch arranca el juego. baseDir es la carpeta del launcher (= carpeta del
// juego). No espera a que el juego termine: lo arranca y devuelve el control.
func Launch(baseDir string, m *manifest.Manifest) error {
	if m.SteamAppID > 0 {
		url := fmt.Sprintf("steam://rungameid/%d", m.SteamAppID)
		log.Printf("[juego] lanzando vía Steam: %s", url)
		return openURL(url)
	}
	return launchExe(baseDir, m)
}

// launchExe arranca el ejecutable del juego indicado en LaunchSpec (plan B si
// no hay AppID de Steam).
func launchExe(baseDir string, m *manifest.Manifest) error {
	spec, ok := m.Launch[runtime.GOOS]
	if !ok || spec.Exe == "" {
		log.Printf("[juego] no hay destino de arranque para %s; nada que lanzar", runtime.GOOS)
		return nil
	}

	exe := spec.Exe
	if !filepath.IsAbs(exe) {
		exe = filepath.Join(baseDir, filepath.FromSlash(spec.Exe))
	}

	log.Printf("[juego] lanzando %s", exe)
	cmd := exec.Command(exe, spec.Args...)
	cmd.Dir = baseDir
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("lanzando el juego: %w", err)
	}
	return nil
}

// openURL abre una URL/protocolo con el manejador del sistema operativo.
func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// El "" es el título de la ventana que exige `start`; evita que
		// interprete la URL como título.
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("abriendo %s: %w", url, err)
	}
	return nil
}
