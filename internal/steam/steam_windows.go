//go:build windows

package steam

import "golang.org/x/sys/windows/registry"

// steamRoots obtiene la(s) carpeta(s) de Steam desde el registro de Windows.
func steamRoots() []string {
	var roots []string
	add := func(root registry.Key, path, name string) {
		key, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
		if err != nil {
			return
		}
		defer key.Close()
		if v, _, err := key.GetStringValue(name); err == nil && v != "" {
			roots = append(roots, v)
		}
	}
	// SteamPath (usuario actual) es lo más fiable; InstallPath como respaldo.
	add(registry.CURRENT_USER, `Software\Valve\Steam`, "SteamPath")
	add(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, "InstallPath")
	add(registry.LOCAL_MACHINE, `SOFTWARE\Valve\Steam`, "InstallPath")
	return roots
}
