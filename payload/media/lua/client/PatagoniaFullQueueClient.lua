-- Patagonia #45 (CLIENTE BASE, distribuido por el launcher) — etiqueta de la cola de servidor-lleno.
--
-- POR QUE VA EN EL CLIENTE BASE Y NO EN EL MOD:
-- La cola de servidor-lleno (#45) se dispara en LoginPacket, en la pantalla de conexion
-- (ConnectToServerState), ANTES de que el cliente descargue y active los mods del server. En ese
-- momento NADA del mod Patagonia esta cargado (ni su Lua ni sus traducciones), asi que un fix dentro
-- del mod nunca corre ahi (mismo fenomeno que nos paso con ItemAwards). Este archivo, al llegar por el
-- launcher, queda en media/lua/client del juego base -> se carga al bootear -> SI esta activo en esa
-- pantalla.
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
