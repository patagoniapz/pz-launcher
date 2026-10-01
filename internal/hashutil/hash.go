// Package hashutil calcula hashes SHA-256 de ficheros locales para comparar
// contra los valores declarados en el manifiesto.
package hashutil

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

// SHA256File devuelve el hash SHA-256 en hexadecimal del fichero indicado.
// Si el fichero no existe devuelve ("", nil) para que el llamante lo trate
// como "hay que descargarlo", en lugar de como un error fatal.
func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
