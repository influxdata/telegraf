#  Minecraft Plugin

This plugin uses the RCON protocol to collect [statistics](http://minecraft.gamepedia.com/Statistics) from a [scoreboard](http://minecraft.gamepedia.com/Scoreboard) on a
Minecraft server.

To enable [RCON](http://wiki.vg/RCON) on the minecraft server, add this to your server configuration in the `server.properties` file:

```
enable-rcon=true
rcon.password=<your password>
rcon.port=<1-65535>
```

To create a new scoreboard objective called `jump` on a minecraft server tracking the `stat.jump` criteria, run this command
in the Minecraft console:

`/scoreboard objectives add jump stat.jump`

Stats are collected with the following RCON command, issued by the plugin:

`/scoreboard players list *`

### Configuration:
```
[[inputs.minecraft]]
   # server address for minecraft
   server = "localhost"
   # port for RCON
   port = "25575"
   # password RCON for mincraft server
   password = "replace_me"
```

### Measurements & Fields:

*This plugin uses only one measurement, titled* `minecraft`

- The field name is the scoreboard objective name.
- The field value is the count of the scoreboard objective

- `minecraft`
    - `<objective_name>` (integer, count)

### Tags:

- The `minecraft` measurement:
    - `server`: the Minecraft RCON server
    - `player`: the Minecraft player


### Sample Queries:

Get the number of jumps per player in the last hour:
```
SELECT SPREAD("jump") FROM "minecraft" WHERE time > now() - 1h GROUP BY "player"
```

### Example Output:

```
$ telegraf --input-filter minecraft --test
* Plugin: inputs.minecraft, Collection 1
> minecraft,player=notch,server=127.0.0.1:25575 jumps=178i 1498261397000000000
> minecraft,player=dinnerbone,server=127.0.0.1:25575 deaths=1i,jumps=1999i,cow_kills=1i 1498261397000000000
> minecraft,player=jeb,server=127.0.0.1:25575 d_pickaxe=1i,damage_dealt=80i,d_sword=2i,hunger=20i,health=20i,kills=1i,level=33i,jumps=264i,armor=15i 1498261397000000000
```
