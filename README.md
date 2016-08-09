HealthD
=================
Run all commands from the `scripts.d`-dir and make the results
available through HTTP.

How does it work?
* (On boot) read all `scripts.d/*.*` entries
* Run these scripts every 5minutes
* Make results available through HTTP (port 10515)

URLS:
*:10515/_mon (plain/text)
*:10515/zenoss (plain/text)
*:10515/health (JSON)

