// Package zomboid localiza la carpeta de datos del juego (logs, partidas…) y
// limpia los logs del cliente.
//
// Ojo: esta carpeta NO es la del launcher. El launcher vive en la carpeta de
// instalación (…\steamapps\common\ProjectZomboid), mientras que los datos del
// usuario están en %UserProfile%\Zomboid (equivalente a ~/Zomboid en Linux/Mac).
package zomboid

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultDir devuelve la ruta estándar de datos de Zomboid: <home>/Zomboid.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Zomboid")
}

// Resolve decide la carpeta de datos de Zomboid con esta prioridad:
//  1. Variable de entorno PZL_ZOMBOID_DIR.
//  2. Valor configurado por el usuario (pzlauncher.json).
//  3. Autodetección (<home>/Zomboid).
func Resolve(configured string) string {
	if env := os.Getenv("PZL_ZOMBOID_DIR"); env != "" {
		return env
	}
	if configured != "" {
		return configured
	}
	return DefaultDir()
}

// Exists indica si la carpeta de datos existe.
func Exists(dir string) bool {
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// CleanLogs borra el contenido de <dir>/Logs y los ficheros de consola de la
// raíz (console.txt y *-console.txt). NO toca partidas, mods ni opciones.
// Devuelve cuántos elementos se eliminaron.
func CleanLogs(dir string) (removed int, err error) {
	// 1. Contenido de la carpeta Logs/ (se conserva la carpeta en sí).
	logsDir := filepath.Join(dir, "Logs")
	if entries, e := os.ReadDir(logsDir); e == nil {
		for _, en := range entries {
			if os.RemoveAll(filepath.Join(logsDir, en.Name())) == nil {
				removed++
			}
		}
	}

	// 2. Logs de consola sueltos en la raíz de Zomboid.
	if entries, e := os.ReadDir(dir); e == nil {
		for _, en := range entries {
			if en.IsDir() {
				continue
			}
			name := en.Name()
			if name == "console.txt" || strings.HasSuffix(name, "-console.txt") {
				if os.Remove(filepath.Join(dir, name)) == nil {
					removed++
				}
			}
		}
	}

	return removed, nil
}
