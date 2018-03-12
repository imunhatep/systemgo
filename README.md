SystemGo is a simple task runner, with configurable jobs with JSON
----

Requires go lang env for compiling.

Features:
 - binary
 - json file configuration
 - task restarting
 - semi-gracefull process closing

```bash
go run main.go -j=2 -f=tasks.json
```

JSON configuration example:
```json
[
  {"name": "php1", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=0", "-x=10"], "restart": true},
  {"name": "php2", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=45", "-x=55"]},
  {"name": "node1", "exec": "/usr/bin/node", "params": ["./gen.js", "-m=0", "-x=10"], "restart": true},
  {"name": "node2", "exec": "/usr/bin/node", "params": ["./gen.js", "-m=45", "-x=55"]}
]

```

CTRL+C to exit process manager.


#### TODO
 - task timeout
 - periodical executor with timer
 - run only once (even if systemg process was terminated)
 - improve logging
 - hot configuration reload
 - task statuses & statistics
 - memory limits
 - web interface (monitoring, stats)
