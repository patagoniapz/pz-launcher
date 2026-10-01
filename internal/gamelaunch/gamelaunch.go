// Package gamelaunch arranca el ejecutable del juego una vez los ficheros están
// actualizados. El objetivo a lanzar se define en el manifiesto por SO, de modo
// que puedas cambiarlo desde el servidor sin recompilar el launcher.
package gamelaunch

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"pzlauncher/internal/manifest"
)

// Launch arranca el juego según el LaunchSpec del SO actual. El ejecutable se
// resuelve relativo a baseDir (la carpeta del launcher). No espera a que el
// juego termine: lo arranca y devuelve el control.
func Launch(baseDir string, m *manifest.Manifest) error {
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
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("lanzando el juego: %w", err)
	}
	return nil
}
