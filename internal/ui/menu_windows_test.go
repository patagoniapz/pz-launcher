//go:build windows

package ui

import "testing"

// TestTaskDialogConfigOffsets verifica que las constantes de desplazamiento de
// TASKDIALOGCONFIG coinciden con el empaquetado real a 1 byte (pshpack1) en x64.
// Reconstruye los offsets sumando los tamaños de los campos en orden de
// declaración y los compara con las constantes que usa el código. Si alguien
// edita un offset por error, este test lo caza sin necesidad de abrir la GUI.
func TestTaskDialogConfigOffsets(t *testing.T) {
	type field struct {
		name string
		size int
	}
	// Campos de TASKDIALOGCONFIG en orden, con su tamaño en x64 (punteros = 8).
	fields := []field{
		{"cbSize", 4},
		{"hwndParent", 8},
		{"hInstance", 8},
		{"dwFlags", 4},
		{"dwCommonButtons", 4},
		{"pszWindowTitle", 8},
		{"mainIcon", 8}, // union HICON/PCWSTR
		{"pszMainInstruction", 8},
		{"pszContent", 8},
		{"cButtons", 4},
		{"pButtons", 8},
		{"nDefaultButton", 4},
		{"cRadioButtons", 4},
		{"pRadioButtons", 8},
		{"nDefaultRadioButton", 4},
		{"pszVerificationText", 8},
		{"pszExpandedInformation", 8},
		{"pszExpandedControlText", 8},
		{"pszCollapsedControlText", 8},
		{"footerIcon", 8}, // union HICON/PCWSTR
		{"pszFooter", 8},
		{"pfCallback", 8},
		{"lpCallbackData", 8},
		{"cxWidth", 4},
	}

	off := map[string]int{}
	n := 0
	for _, f := range fields {
		off[f.name] = n
		n += f.size
	}

	if n != tdcSize {
		t.Fatalf("tamaño total de TASKDIALOGCONFIG = %d, esperaba %d", n, tdcSize)
	}

	checks := []struct {
		name string
		got  int
	}{
		{"cbSize", offCbSize},
		{"dwFlags", offDwFlags},
		{"pszWindowTitle", offPszWindowTitle},
		{"pszMainInstruction", offPszMainInstruction},
		{"pszContent", offPszContent},
		{"cButtons", offCButtons},
		{"pButtons", offPButtons},
		{"nDefaultButton", offNDefaultButton},
	}
	for _, c := range checks {
		if off[c.name] != c.got {
			t.Errorf("offset de %s = %d (constante), empaquetado real = %d", c.name, c.got, off[c.name])
		}
	}

	// TASKDIALOG_BUTTON: int (4) + PCWSTR (8) empaquetado = 12 bytes.
	if tdbSize != 12 {
		t.Errorf("tdbSize = %d, esperaba 12", tdbSize)
	}
}
