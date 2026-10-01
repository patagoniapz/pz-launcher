// Package paths resuelve rutas relativas a la ubicación del propio ejecutable,
// no al directorio de trabajo. Esto permite dejar el launcher dentro de la
// carpeta del juego (p. ej. ...\Steam\steamapps\common\ProjectZomboid) y que
// todas las rutas se calculen a partir de ahí, se ejecute desde donde se ejecute.
package paths

import (
	"os"
	"path/filepath"
)

// ExecutableDir devuelve la carpeta que contiene al ejecutable en marcha,
// resolviendo symlinks (p. ej. si se invoca a través de un enlace).
func ExecutableDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Dir(exe), nil
}
