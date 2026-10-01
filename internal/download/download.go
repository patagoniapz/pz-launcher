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
//
// onProgress, si no es nil, se llama durante la descarga con los bytes recibidos
// y el total esperado (total es -1 si el servidor no lo informa).
func Verified(url, dir, wantHash string, onProgress func(done, total int64)) (string, error) {
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
	dst := io.Writer(io.MultiWriter(tmp, h))
	if onProgress != nil {
		dst = io.MultiWriter(tmp, h, &progressWriter{total: resp.ContentLength, onProgress: onProgress})
	}
	if _, err := io.Copy(dst, resp.Body); err != nil {
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

// progressWriter cuenta los bytes escritos y los reporta vía onProgress.
type progressWriter struct {
	done       int64
	total      int64
	onProgress func(done, total int64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.done += int64(n)
	p.onProgress(p.done, p.total)
	return n, nil
}
