-- Patagonia #45 — etiqueta de la cola de servidor-lleno (pantalla de conexion).
--
-- ESTE ARCHIVO VIVE EN DOS LUGARES IDENTICOS, CADA UNO CON SU ROL:
--   1) Cliente BASE via launcher (launcher/payload/media/lua/client/) -> se carga al bootear, asi
--      esta ACTIVO en la pantalla de conexion (ConnectToServerState), que corre ANTES de que el
--      cliente descargue/active los mods del server. Por eso NO alcanza con tenerlo solo en el mod:
--      el Lua del mod no esta cargado todavia en esa pantalla (mismo fenomeno que ItemAwards).
--   2) Mod Patagonia (media/lua/client/) -> su UNICO proposito aqui es que el SERVIDOR tenga el
--      archivo automaticamente (via el deploy normal del mod), para que pase el file-check de PZ
--      ("File doesn't exist on the server: media/lua/client/..."). En el server dedicado es inerte
--      (headless no ejecuta Lua cliente); solo tiene que existir en la misma ruta virtual.
-- MANTENER AMBAS COPIAS IDENTICAS (mismo contenido) para que coincida la ruta/los bytes.
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
