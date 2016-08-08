HealthD
=================
Run all commands from the `conf.d`-dir and make the results
available through HTTP.

Example conf.d
```
cmd = "/usr/bin/php"
args = "test.php"
# prefix added to string on error
errprefix = "example"
```

How does it work?
* (On boot) read all `conf.d/*.toml` entries
* Run conf.d commands every 5minutes
* Make results available through *:10515/zenoss (plain/text) and *:10515/health (JSON)
