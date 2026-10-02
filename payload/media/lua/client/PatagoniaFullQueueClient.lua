-- Patagonia #45 — etiqueta de la cola de servidor-lleno (pantalla de conexion).
--
-- ESTE ARCHIVO ES DEL CLIENTE BASE (media/lua/client), NO va en el mod. Razones:
--   A) CARGA: la pantalla de conexion (ConnectToServerState) corre ANTES de que el cliente
--      descargue/active los mods del server -> un Lua dentro del mod no esta cargado ahi (mismo
--      fenomeno que ItemAwards). Como archivo base se carga al bootear y SI esta activo en esa pantalla.
--   B) CHECKSUM (NetChecksum/LuaManager.LoadDirBase): el server valida el Lua del cliente comparando
--      la lista de archivos POR POSICION. Esa lista = [archivos BASE ordenados] + [archivos de MODS al
--      final] (con dedup por ruta, gana el base). Si este archivo estuviera en base en el cliente pero
--      en el mod en el server, quedaria en secciones distintas -> listas DESALINEADAS -> el server
--      reporta falsos "File doesn't exist on the client" en archivos vanilla siguientes (p.ej.
--      pzapi/ModOptions.lua). Por eso DEBE estar en media/lua/client BASE EN AMBOS LADOS.
--
-- DISTRIBUCION:
--   * Cliente: lo instala el launcher (este payload) en <juego>/media/lua/client/.
--   * Servidor: hay que copiar el MISMO archivo (byte-identico) al <server>/media/lua/client/ en el
--     deploy, junto al jar #45. NO alcanza con tenerlo en el mod (ver punto B). Tras copiarlo hay que
--     REINICIAR el server para que recompute GameServer.checksum incluyendolo.
--
-- QUE HACE:
-- Cuando el server esta lleno, el jar #45 encola al jugador y le manda QueuePacket PlaceInQueue con su
-- posicion. El cliente dispara OnConnectionStateChanged("FormatMessage", "PlaceInQueue", n), pero el
-- vanilla lo formatea con la clave 'UI_servers_PlaceInQueue', que NO existe -> saldria la clave cruda.
-- Reescribimos la etiqueta con una clave que SI existe en todos los idiomas vanilla (la usa
-- LoadingQueueUI), y que ademas muestra el NUMERO de posicion:
--   UI_GameLoad_PlaceInQueue = "Ocupas la posicion %1 en la cola de conexion"
-- getText YA formatea el token %1 (NO usar string.format; ver gotcha pz-translator-format-percent).
--
-- Sin el jar #45 en el server esto es INOFENSIVO: el evento PlaceInQueue en fase de conexion solo
-- ocurre si el server encola, cosa que solo hace el jar parcheado. Puro cliente, no toca red.

local function onConnectionStateChanged(state, message, arg)
    if state ~= "FormatMessage" or message ~= "PlaceInQueue" then
        return
    end
    local cts = ConnectToServer and ConnectToServer.instance
    if not cts or not cts.getIsVisible or not cts:getIsVisible() then
        return  -- durante la cola de CARGA (LoadingQueueState) esta pantalla no es visible: no tocar
    end
    if cts.connectLabel then
        cts.connectLabel.name = getText("UI_GameLoad_PlaceInQueue", arg)
    end
end

Events.OnConnectionStateChanged.Add(onConnectionStateChanged)
