// Package manifest define el formato del manifest.json y su descarga.
//
// El manifiesto es el "contrato" entre el servidor (GitHub Releases) y el
// launcher: describe qué ficheros existen, su hash SHA-256 esperado, de dónde
// bajarlos, qué versión del propio launcher es la actual y cómo arrancar el
// juego en cada plataforma.
package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Manifest es el documento raíz publicado en el release.
type Manifest struct {
	// Schema versiona el formato del propio manifiesto por si evoluciona.
	Schema int `json:"schema"`

	// LauncherVersion es la última versión disponible del launcher (semver).
	LauncherVersion string `json:"launcher_version"`

	// LauncherBinaries mapea "GOOS/GOARCH" -> binario del launcher, para la
	// autoactualización. Ej.: "windows/amd64", "linux/amd64", "darwin/arm64".
	LauncherBinaries map[string]Asset `json:"launcher_binaries"`

	// Files son los ficheros del juego gestionados (p. ej. projectzomboid.jar).
	Files []File `json:"files"`

	// Launch indica cómo arrancar el juego por SO (clave = runtime.GOOS).
	Launch map[string]LaunchSpec `json:"launch"`

	// SteamAppID, si es > 0, hace que el juego se lance vía Steam
	// (steam://rungameid/<id>), lo que arranca Steam si está cerrado.
	// Project Zomboid es 108600. Si es 0, se usa el ejecutable de Launch.
	SteamAppID int `json:"steam_appid,omitempty"`

	// PatchNotes son las novedades/parches aplicados (en español), para mostrar
	// al jugador desde el menú "Novedades". Convención: el más nuevo primero.
	// Viajan en el manifiesto para poder actualizarlas sin recompilar el launcher.
	PatchNotes []PatchNote `json:"patch_notes,omitempty"`
}

// PatchNote describe un parche/novedad aplicado al juego, tal como lo ve el
// jugador. Los textos están en español (es el idioma del usuario final).
type PatchNote struct {
	// ID es el identificador corto del parche, p. ej. "C13". Opcional.
	ID string `json:"id,omitempty"`
	// Title es un título breve de una línea.
	Title string `json:"title"`
	// Desc amplía qué cambia o qué arregla (puede ocupar varias líneas). Opcional.
	Desc string `json:"desc,omitempty"`
	// Date es la fecha del parche (YYYY-MM-DD). Opcional.
	Date string `json:"date,omitempty"`
}

// Asset es un fichero descargable con su hash de verificación.
type Asset struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// File es un fichero del juego gestionado por el launcher.
type File struct {
	// Path es relativo a la carpeta del ejecutable, con separador "/".
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	URL    string `json:"url"`
}

// LaunchSpec describe el ejecutable del juego a lanzar tras actualizar.
type LaunchSpec struct {
	// Exe es relativo a la carpeta del launcher (p. ej. "ProjectZomboid64.exe").
	Exe  string   `json:"exe"`
	Args []string `json:"args"`
}

// Fetch descarga y decodifica el manifiesto desde una URL.
//
// La petición rompe cachés intermedias (CDN/proxies): justo después de publicar
// un release, una copia cacheada del manifiesto haría que el launcher validara
// contra los hashes VIEJOS y dijera "todo al día" cuando el servidor ya exige el
// jar nuevo. Un parámetro único por petición cambia la clave de caché, y las
// cabeceras piden contenido fresco a las cachés que las respetan.
func Fetch(url string) (*Manifest, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	req, err := http.NewRequest(http.MethodGet, url+sep+"ts="+strconv.FormatInt(time.Now().UnixNano(), 10), nil)
	if err != nil {
		return nil, fmt.Errorf("preparando petición del manifiesto: %w", err)
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("descargando manifiesto: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifiesto devolvió HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20)) // 5 MiB de tope
	if err != nil {
		return nil, fmt.Errorf("leyendo manifiesto: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parseando manifiesto: %w", err)
	}
	return &m, nil
}
