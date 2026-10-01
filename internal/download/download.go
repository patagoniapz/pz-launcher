// Package download descarga ficheros a una ruta temporal verificando su
// hash SHA-256 ANTES de que el llamante los mueva a su destino final.
//
// Descargar a un temporal y verificar primero evita dejar un fichero del juego
// a medio escribir o corrupto si la descarga falla o la red se corta.
package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Verified descarga url en un fichero temporal dentro de dir, comprueba que su
// SHA-256 coincide con wantHash y devuelve la ruta del temporal. El llamante es
// responsable de moverlo a su destino (os.Rename) y de borrarlo si algo falla.
func Verified(url, dir, wantHash string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("descargando %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("descargando %s: HTTP %d", url, resp.StatusCode)
	}

	tmp, err := os.CreateTemp(dir, ".pzl-download-*")
	if err != nil {
		return "", fmt.Errorf("creando temporal: %w", err)
	}
	tmpPath := tmp.Name()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("guardando descarga: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("cerrando temporal: %w", err)
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, wantHash) {
		os.Remove(tmpPath)
		return "", fmt.Errorf("hash no coincide en %s: esperado %s, obtenido %s",
			filepath.Base(url), wantHash, got)
	}

	return tmpPath, nil
}
