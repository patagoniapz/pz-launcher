// Package selfupdate actualiza el propio ejecutable del launcher.
//
// Un programa no puede borrar su propio .exe mientras se ejecuta en Windows,
// pero SÍ puede renombrarlo. El truco estándar es:
//
//  1. Descargar el nuevo binario a "<exe>.new" (verificando hash).
//  2. Renombrar el actual a "<exe>.old".
//  3. Renombrar "<exe>.new" al nombre real.
//  4. Relanzar el nuevo binario y salir.
//
// Al arrancar, CleanupOld() borra el "<exe>.old" que dejó la ejecución anterior.
// En Linux/macOS este baile también funciona (se puede sustituir un binario en
// ejecución porque el inodo abierto sigue vivo).
package selfupdate

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"pzlauncher/internal/download"
	"pzlauncher/internal/manifest"
)

// MaybeUpdate comprueba si el manifiesto anuncia una versión del launcher más
// nueva que current. Si es así, la descarga, la instala y relanza el proceso.
// Devuelve relaunched=true cuando ha arrancado el nuevo binario (el llamante
// debe terminar inmediatamente en ese caso).
func MaybeUpdate(current string, m *manifest.Manifest) (relaunched bool, err error) {
	if m.LauncherVersion == "" || !isNewer(m.LauncherVersion, current) {
		return false, nil
	}

	key := runtime.GOOS + "/" + runtime.GOARCH
	asset, ok := m.LauncherBinaries[key]
	if !ok || asset.URL == "" {
		log.Printf("[launcher] hay versión %s pero no hay binario para %s; se omite autoactualización",
			m.LauncherVersion, key)
		return false, nil
	}

	log.Printf("[launcher] actualizando %s -> %s", current, m.LauncherVersion)

	exe, err := os.Executable()
	if err != nil {
		return false, err
	}
	exe, _ = filepath.EvalSymlinks(exe)
	dir := filepath.Dir(exe)

	tmpPath, err := download.Verified(asset.URL, dir, asset.SHA256)
	if err != nil {
		return false, err
	}

	oldPath := exe + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(exe, oldPath); err != nil {
		os.Remove(tmpPath)
		return false, fmt.Errorf("renombrando launcher actual: %w", err)
	}
	if err := os.Rename(tmpPath, exe); err != nil {
		_ = os.Rename(oldPath, exe) // restaurar
		os.Remove(tmpPath)
		return false, fmt.Errorf("instalando nuevo launcher: %w", err)
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(exe, 0o755)
	}

	// Relanzar el nuevo binario con los mismos argumentos.
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("relanzando launcher: %w", err)
	}
	return true, nil
}

// CleanupOld borra el "<exe>.old" dejado por una autoactualización previa.
// Se llama al arrancar; si el .old sigue bloqueado se ignora silenciosamente.
func CleanupOld() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exe, _ = filepath.EvalSymlinks(exe)
	_ = os.Remove(exe + ".old")
}

// isNewer devuelve true si a es una versión semver estrictamente mayor que b.
// Comparación tolerante: ignora prefijo "v" y partes no numéricas.
func isNewer(a, b string) bool {
	pa, pb := parseVersion(a), parseVersion(b)
	for i := 0; i < len(pa) && i < len(pb); i++ {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return len(pa) > len(pb)
}

func parseVersion(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		// Cortar cualquier sufijo tipo "-beta".
		if i := strings.IndexAny(p, "-+"); i >= 0 {
			p = p[:i]
		}
		n, _ := strconv.Atoi(p)
		out = append(out, n)
	}
	return out
}
