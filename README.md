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
  {"name": "php1", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=0", "-x=10"], "restart": 0},
  {"name": "php2", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=15", "-x=25"]},
  {"name": "node1", "exec": "/usr/bin/node", "params": ["./gen.js", "-m=30", "-x=40"], "restart": 5},
  {"name": "node2", "exec": "/usr/bin/node", "params": ["./gen.js", "-m=45", "-x=55"]}
]
```

*restart* - seconds between job restart (after finishing). O (zero) means - do not restart.


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
