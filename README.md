# Reforger Server Admin Tools Exporter

This program will read the stats file written by [Server Admin
Tools](https://reforger.armaplatform.com/workshop/5AAAC70D754245DD-ServerAdminTools)
and export the statistics in a Prometheus-compatible metrics format.

The tool is meant to run alongside the server, and will by default export the
metrics on port 9840. It is recommended to run the tools as an unprivileged
user, but naturally it must be able to read the file created by SAT.

The command line options are:

```
Usage of sat_exporter:
  -f string
        file to read stats from (default "/home/reforger/profile/profile/ServerAdminTools_Stats.json")
  -l string
        Labels/values to augment metrics with, in the form label1=val1,label2=val2
  -listen string
        ip:port to listen on (default ":9840")
  -namespace string
        Namespace (prefix) to use for Prometheus metrics (default "reforger_sat_exporter")
  -once
        Only output the stats to stdout and exit (for testing)
  -timeout duration
        Timeout for webserver reading client request (default 3s)
```

They should be pretty self-explanatory, and usually you only need to use `-f`.

The option `-once` is useful to see whether the tool can read your file, and if
it outputs credible numbers.

Example metrics (eliding the Go runtime metrics autogenerated by the prom client):

```
# HELP reforger_sat_exporter_ai_characters Total number AI characters
# TYPE reforger_sat_exporter_ai_characters gauge
reforger_sat_exporter_ai_characters 0
# HELP reforger_sat_exporter_frames_per_second Frames per second server-side
# TYPE reforger_sat_exporter_frames_per_second gauge
reforger_sat_exporter_frames_per_second 89
# HELP reforger_sat_exporter_player_count Current number of players
# TYPE reforger_sat_exporter_player_count gauge
reforger_sat_exporter_player_count 0
# HELP reforger_sat_exporter_registered_entities Total number of registered entities
# TYPE reforger_sat_exporter_registered_entities gauge
reforger_sat_exporter_registered_entities 973
# HELP reforger_sat_exporter_registered_groups Total number of registered groups
# TYPE reforger_sat_exporter_registered_groups gauge
reforger_sat_exporter_registered_groups 0
# HELP reforger_sat_exporter_registered_systems Total number of registered systems
# TYPE reforger_sat_exporter_registered_systems gauge
reforger_sat_exporter_registered_systems 83
# HELP reforger_sat_exporter_registered_tasks Total number registered tasks
# TYPE reforger_sat_exporter_registered_tasks gauge
reforger_sat_exporter_registered_tasks 0
# HELP reforger_sat_exporter_uptime_seconds Server uptime in seconds
# TYPE reforger_sat_exporter_uptime_seconds counter
reforger_sat_exporter_uptime_seconds 3942
```

The file `dashboard.json` contains an example Grafana dashboard.
