#  Minecraft Plugin

## Configuration:
```
[[inputs.minecraft]]
   # server address for minecraft
   server = "localhost"
   # port for RCON
   port = "25575"
   # password RCON for mincraft server
   password = "replace_me"
```

## Description

This plugin uses the RCON protocol to collect [statistics](http://minecraft.gamepedia.com/Statistics) from a [scoreboard](http://minecraft.gamepedia.com/Scoreboard) on a
Minecraft server.

To enable [RCON](http://wiki.vg/RCON) on the minecraft server, add this to your server configuration:

```
  enable-rcon=true
  rcon.password=<your password>
  rcon.port=<1-65535>
```

To create a new scoreboard objective called `jump` on a minecraft server tracking the `jump` stat, run this command
in the Minecraft console:

`/scoreboard objectives add jump stat.jump`

Stats are collected with the following RCON command, issued by the plugin:

`scoreboard players list *`

## Measurements:
### Minecraft measurement

*This plugin uses only one measurement, titled* `minecraft`


### Tags:

- The `minecraft` measurement:
    - `server`: the Minecraft RCON server
    - `player`: the Minecraft player



### Fields:
- The field name is the scoreboard objective name.
- The field value is the count of the scoreboard objective
