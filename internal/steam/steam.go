// Package steam localiza la carpeta de INSTALACIÓN de un juego de Steam
// (donde está projectzomboid.jar), no la de datos del usuario (logs/partidas).
//
// Steam no expone una variable de entorno con esa ruta, pero se puede deducir:
//  1. La carpeta de Steam (según el SO: registro en Windows, rutas conocidas en
//     Linux/Mac).
//  2. Las bibliotecas (posiblemente en varios discos) desde libraryfolders.vdf.
//  3. En cada biblioteca, appmanifest_<appid>.acf da el "installdir"; la carpeta
//     final es <biblioteca>/steamapps/common/<installdir>.
package steam

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	reInstallDir = regexp.MustCompile(`(?i)"installdir"\s+"([^"]+)"`)
	rePath       = regexp.MustCompile(`(?i)"path"\s+"([^"]+)"`)
)

// FindGameDir devuelve la carpeta de instalación del juego con el AppID dado
// (Project Zomboid = 108600) y true si la encontró.
func FindGameDir(appID int) (string, bool) {
	seen := map[string]bool{}
	for _, root := range steamRoots() {
		for _, lib := range libraryPaths(root) {
			if seen[lib] {
				continue
			}
			seen[lib] = true
			if dir, ok := gameInLibrary(lib, appID); ok {
				return dir, true
			}
		}
	}
	return "", false
}

// gameInLibrary mira si una biblioteca contiene el juego y devuelve su carpeta.
func gameInLibrary(lib string, appID int) (string, bool) {
	acf := filepath.Join(lib, "steamapps", "appmanifest_"+strconv.Itoa(appID)+".acf")
	data, err := os.ReadFile(acf)
	if err != nil {
		return "", false
	}
	m := reInstallDir.FindSubmatch(data)
	if m == nil {
		return "", false
	}
	dir := filepath.Join(lib, "steamapps", "common", string(m[1]))
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir, true
	}
	return "", false
}

// libraryPaths lee las bibliotecas declaradas (la propia carpeta de Steam más
// las de libraryfolders.vdf, que pueden estar en otros discos).
func libraryPaths(steamRoot string) []string {
	libs := []string{steamRoot}
	candidates := []string{
		filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf"),
		filepath.Join(steamRoot, "config", "libraryfolders.vdf"),
	}
	for _, vdf := range candidates {
		data, err := os.ReadFile(vdf)
		if err != nil {
			continue
		}
		for _, m := range rePath.FindAllSubmatch(data, -1) {
			// En el VDF las barras invertidas van escapadas ("\\").
			libs = append(libs, strings.ReplaceAll(string(m[1]), `\\`, `\`))
		}
	}
	return libs
}
