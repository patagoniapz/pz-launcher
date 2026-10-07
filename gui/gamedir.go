package main

import (
	"errors"
	"os"
	"path/filepath"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"pzlauncher/internal/config"
	"pzlauncher/internal/steam"
	"pzlauncher/internal/zomboid"
)

// resolveGameDir localiza la carpeta de instalación del juego con la misma
// prioridad que cmd/launcher, pero usando el selector de carpeta nativo de
// Wails en vez del de zenity:
//  1. Variable de entorno PZL_GAME_DIR.
//  2. Valor guardado en la config (pzlauncher.json).
//  3. El propio exe ya está dentro de la carpeta del juego.
//  4. Autodetección vía Steam.
//  5. Preguntar al usuario (y recordar la elección).
func (a *App) resolveGameDir(exeDir string, appID int) (string, error) {
	if env := os.Getenv("PZL_GAME_DIR"); env != "" {
		return env, nil
	}
	a.mu.Lock()
	saved := a.cfg.GameDir
	a.mu.Unlock()
	if saved != "" && dirExists(saved) {
		return saved, nil
	}
	if hasGameExe(exeDir) {
		return exeDir, nil
	}
	if dir, ok := steam.FindGameDir(appID); ok {
		return dir, nil
	}
	picked, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Selecciona la carpeta de instalación de Project Zomboid",
	})
	if err == nil && picked != "" {
		a.mu.Lock()
		a.cfg.GameDir = picked
		cfg := a.cfg
		exe := a.exeDir
		a.mu.Unlock()
		_ = config.Save(exe, cfg)
		return picked, nil
	}
	return "", errors.New("No se encontró la carpeta de instalación de Project Zomboid.")
}

// resolveZomboidDir localiza la carpeta de DATOS de Zomboid (logs, partidas…),
// preguntando con el selector nativo si no se autodetecta.
func (a *App) resolveZomboidDir() (string, bool) {
	a.mu.Lock()
	zdir := a.cfg.ZomboidDir
	a.mu.Unlock()
	dir := zomboid.Resolve(zdir)
	if zomboid.Exists(dir) {
		return dir, true
	}
	picked, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Selecciona tu carpeta Zomboid (donde están los logs)",
	})
	if err != nil || picked == "" {
		return "", false
	}
	a.mu.Lock()
	a.cfg.ZomboidDir = picked
	cfg := a.cfg
	exe := a.exeDir
	a.mu.Unlock()
	_ = config.Save(exe, cfg)
	return picked, true
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
