# pzlauncher

> 🇬🇧 English below · 🇪🇸 Español más abajo

---

## 🇬🇧 English

Cross-platform **launcher / updater for Project Zomboid**. A single executable
you can run from **anywhere** — it **auto-detects your Steam install** of the
game (registry + `libraryfolders.vdf`, so it also finds installs on other
drives). On launch it:

1. Reads a `manifest.json` from GitHub Releases.
2. Self-updates if a newer launcher version is published.
3. Compares each managed file (e.g. `projectzomboid.jar`) by **SHA-256** and
   downloads only what changed.
4. Starts the game.

Progress is shown in a small **native window** with an **estimated time left**
(falls back to console if no GUI is available). On Windows it carries the
Project Zomboid icon.

It does **not** auto-launch: after updating it asks **Play / Clean logs / Quit**.
"Clean logs" only cleans (it does not start the game) and re-shows the dialog.
The game is started **through Steam** (`steam://rungameid/108600`), which boots
Steam if it is closed. "Clean logs" wipes `%UserProfile%\Zomboid\Logs\` and the
`*console.txt` files (never your saves or mods).

Downloads are resilient to flaky connections: a stalled transfer is detected and
automatically retried.

The **game install** folder is auto-detected via Steam; the **Zomboid data**
folder (logs/saves) is autodetected at `%UserProfile%\Zomboid` (`~/Zomboid`).
Both can be overridden with env vars (`PZL_GAME_DIR`, `PZL_ZOMBOID_DIR`) or a
`pzlauncher.json` next to the executable:
`{ "game_dir": "D:\\...\\ProjectZomboid", "zomboid_dir": "D:\\...\\Zomboid" }`.
Run `pzlauncher --detect` to print the detected game folder.

### Get it

A prebuilt binary is provided in the [Releases](../../releases) section for
Windows, Linux and macOS — just download and drop it in the game folder.

If you prefer, you can **download the source and build it yourself**:

```bash
# requires Go 1.25+
git clone https://github.com/patagoniapz/pz-launcher.git
cd pz-launcher
go build -o pzlauncher ./cmd/launcher     # add .exe on Windows
```

Optionally pin version and manifest URL:

```bash
go build -ldflags "-X main.version=1.0.0 -X main.manifestURL=https://github.com/patagoniapz/pz-launcher/releases/latest/download/manifest.json" -o pzlauncher ./cmd/launcher
```

---

## 🇪🇸 Español

**Launcher / actualizador multiplataforma para Project Zomboid**. Un único
ejecutable que puedes ejecutar desde **cualquier sitio** — **autodetecta tu
instalación de Steam** del juego (registro + `libraryfolders.vdf`, así encuentra
también instalaciones en otros discos). Al arrancar:

1. Lee un `manifest.json` desde GitHub Releases.
2. Se autoactualiza si hay una versión más nueva del launcher.
3. Compara cada fichero gestionado (p. ej. `projectzomboid.jar`) por **SHA-256**
   y descarga solo lo que haya cambiado.
4. Arranca el juego.

El progreso se muestra en una pequeña **ventana nativa** con **tiempo estimado
restante** (si no hay entorno gráfico, cae a consola). En Windows lleva el icono
de Project Zomboid.

**No** arranca el juego solo: tras actualizar pregunta **Jugar / Limpiar logs /
Salir**. "Limpiar logs" solo limpia (no arranca el juego) y vuelve a mostrar el
diálogo. El juego se lanza **vía Steam** (`steam://rungameid/108600`), que abre
Steam si está cerrado. "Limpiar logs" borra `%UserProfile%\Zomboid\Logs\` y los
`*console.txt` (nunca tus partidas ni mods).

Las descargas toleran conexiones inestables: si la transferencia se queda parada,
se detecta y se reintenta automáticamente.

La carpeta de **instalación** del juego se autodetecta vía Steam; la carpeta de
**datos** de Zomboid (logs/partidas) se autodetecta en `%UserProfile%\Zomboid`
(`~/Zomboid`). Ambas se pueden cambiar con variables (`PZL_GAME_DIR`,
`PZL_ZOMBOID_DIR`) o con un `pzlauncher.json` junto al ejecutable:
`{ "game_dir": "D:\\...\\ProjectZomboid", "zomboid_dir": "D:\\...\\Zomboid" }`.
Ejecuta `pzlauncher --detect` para ver la carpeta del juego detectada.

### Cómo obtenerlo

Se ofrece un binario ya compilado en la sección de [Releases](../../releases)
para Windows, Linux y macOS — basta con descargarlo y dejarlo en la carpeta del
juego.

Si lo prefieres, puedes **descargar el código fuente y compilarlo por tus
propios medios**:

```bash
# requiere Go 1.22+
git clone https://github.com/patagoniapz/pz-launcher.git
cd pz-launcher
go build -o pzlauncher ./cmd/launcher     # añade .exe en Windows
```

Opcionalmente, fija la versión y la URL del manifiesto:

```bash
go build -ldflags "-X main.version=1.0.0 -X main.manifestURL=https://github.com/patagoniapz/pz-launcher/releases/latest/download/manifest.json" -o pzlauncher ./cmd/launcher
```
