Isolated plugin

1. Start telegraf
2. For each plugin, get the name of the plugin and pass it to execd along with the config path
    `telegraf plugin $plugin-name $config-path`
3. When telegraf is called with `plugin` command it will only launch the plugin with the given name
