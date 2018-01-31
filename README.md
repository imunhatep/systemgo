SystemG is a simple task runner, with configurable jobs with JSON
----

Requires go lang env for compiling.

Features:
 - binary
 - json file configuration
 - task restarting
 - semi-gracefull process closing

```bash
go run main.go
```

JSON configuration example:
```json
[
  {"name": "php1", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=0", "-x=20"], "restart": true},
  {"name": "php2", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]},
  {"name": "js1", "exec": "/usr/bin/node", "params": ["./gen.js", "-m=50", "-x=60"]}
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
 
  