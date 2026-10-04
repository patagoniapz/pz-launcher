//go:build !windows

package ui

// showMenu en Linux/macOS usa la lista nativa de zenity (no hay TaskDialog de
// Windows). def es la acción preseleccionada por defecto.
func showMenu(instruction, content string, items []menuItem, def Action) Action {
	return zenityListMenu(instruction, content, items, def)
}
