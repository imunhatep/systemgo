Process manager (systemG).
======

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
  {"name": "php3", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]},
  {"name": "php4", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]},
  {"name": "php5", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]},
  {"name": "php6", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]},
  {"name": "php7", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]},
  {"name": "php8", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]},
  {"name": "php9", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]},
  {"name": "php10", "exec": "/usr/bin/php", "params": ["./gen.php", "-m=50", "-x=60"]}
]
```

CTRL+C to exit process manager.