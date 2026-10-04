//go:build windows

// En Windows el menú principal se dibuja con un TaskDialog nativo (comctl32):
// una lista vertical de "command links" (botones grandes), uno por acción. Es
// más intuitivo que una lista y, al ir el texto en el área de contenido, se
// ajusta en varias líneas en vez de recortarse.
//
// TaskDialog necesita Common Controls v6; el launcher ya lo activa a nivel de
// proceso al abrir el primer diálogo de zenity (ventana de progreso), que fija
// el contexto de activación por defecto. Si por lo que sea TaskDialogIndirect no
// está disponible o falla, caemos a la lista de zenity (zenityListMenu).
package ui

import (
	"encoding/binary"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modComctl32            = windows.NewLazySystemDLL("comctl32.dll")
	procTaskDialogIndirect = modComctl32.NewProc("TaskDialogIndirect")
)

// Flags de TASKDIALOGCONFIG (dwFlags) que usamos.
const (
	tdfAllowDialogCancellation  = 0x0008 // permite cerrar con la X / Esc
	tdfUseCommandLinks          = 0x0010 // botones grandes verticales
	tdfPositionRelativeToWindow = 0x1000
)

// idCancel es el ID que devuelve TaskDialog al cerrar con la X o Esc (IDCANCEL).
const idCancel = 2

// idBase es el primer ID de botón personalizado. Empieza alto para no chocar con
// los IDs estándar (IDOK=1, IDCANCEL=2, …).
const idBase = 100

// Desplazamientos de los campos dentro de TASKDIALOGCONFIG, que está empaquetado
// a 1 byte (pshpack1 en commctrl.h) en x64. Verificados en menu_windows_test.go.
const (
	tdcSize               = 160 // cbSize de TASKDIALOGCONFIG
	offCbSize             = 0
	offDwFlags            = 20
	offPszWindowTitle     = 28
	offPszMainInstruction = 44
	offPszContent         = 52
	offCButtons           = 60
	offPButtons           = 64
	offNDefaultButton     = 72
)

// tdbSize es el tamaño de TASKDIALOG_BUTTON empaquetado: int nButtonID (4) +
// PCWSTR pszButtonText (8) = 12 bytes.
const tdbSize = 12

// showMenu dibuja el menú con TaskDialog; si no está disponible o falla, usa la
// lista de zenity como plan B.
func showMenu(instruction, content string, items []menuItem, def Action) Action {
	if act, ok := taskDialogMenu(instruction, content, items, def); ok {
		return act
	}
	return zenityListMenu(instruction, content, items, def)
}

func taskDialogMenu(instruction, content string, items []menuItem, def Action) (Action, bool) {
	if err := procTaskDialogIndirect.Find(); err != nil {
		return 0, false
	}

	// Array TASKDIALOG_BUTTON (empaquetado). texts mantiene vivos los punteros a
	// las cadenas UTF-16 referenciadas desde el buffer.
	btnBuf := make([]byte, tdbSize*len(items))
	texts := make([][]uint16, len(items))
	defID := idBase
	for i, it := range items {
		id := idBase + i
		if it.action == def {
			defID = id
		}
		t, err := syscall.UTF16FromString(it.label)
		if err != nil {
			return 0, false
		}
		texts[i] = t
		base := i * tdbSize
		binary.LittleEndian.PutUint32(btnBuf[base:], uint32(id))
		binary.LittleEndian.PutUint64(btnBuf[base+4:], uint64(uintptr(unsafe.Pointer(&texts[i][0]))))
	}

	title := mustUTF16(appTitle)
	instr := mustUTF16(instruction)
	var contentPtr unsafe.Pointer
	var cont []uint16
	if content != "" {
		cont = mustUTF16(content)
		contentPtr = unsafe.Pointer(&cont[0])
	}

	cfgBuf := make([]byte, tdcSize)
	putU32 := func(off int, v uint32) { binary.LittleEndian.PutUint32(cfgBuf[off:], v) }
	putPtr := func(off int, p unsafe.Pointer) {
		binary.LittleEndian.PutUint64(cfgBuf[off:], uint64(uintptr(p)))
	}
	putU32(offCbSize, tdcSize)
	putU32(offDwFlags, tdfUseCommandLinks|tdfAllowDialogCancellation|tdfPositionRelativeToWindow)
	putPtr(offPszWindowTitle, unsafe.Pointer(&title[0]))
	putPtr(offPszMainInstruction, unsafe.Pointer(&instr[0]))
	putPtr(offPszContent, contentPtr)
	putU32(offCButtons, uint32(len(items)))
	putPtr(offPButtons, unsafe.Pointer(&btnBuf[0]))
	putU32(offNDefaultButton, uint32(int32(defID)))

	// TaskDialog corre su propio bucle de mensajes modal: fijamos el hilo del SO.
	runtime.LockOSThread()
	var pnButton int32
	ret, _, _ := procTaskDialogIndirect.Call(
		uintptr(unsafe.Pointer(&cfgBuf[0])),
		uintptr(unsafe.Pointer(&pnButton)),
		0, // pnRadioButton (sin radios)
		0, // pfVerificationFlagChecked (sin checkbox)
	)
	runtime.UnlockOSThread()

	// Mantén vivos todos los datos apuntados hasta que la llamada termina.
	runtime.KeepAlive(cfgBuf)
	runtime.KeepAlive(btnBuf)
	runtime.KeepAlive(texts)
	runtime.KeepAlive(title)
	runtime.KeepAlive(instr)
	runtime.KeepAlive(cont)

	if int32(ret) != 0 { // HRESULT != S_OK
		return 0, false
	}
	if pnButton == idCancel {
		return ActionQuit, true
	}
	idx := int(pnButton) - idBase
	if idx >= 0 && idx < len(items) {
		return items[idx].action, true
	}
	return ActionQuit, true
}

// mustUTF16 convierte s a UTF-16 terminado en NUL. Las etiquetas del menú no
// contienen NUL, así que el error de UTF16FromString no puede darse aquí.
func mustUTF16(s string) []uint16 {
	v, err := syscall.UTF16FromString(s)
	if err != nil {
		return []uint16{0}
	}
	return v
}
