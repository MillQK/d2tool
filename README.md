This CLI collects stats of dota heroes and creates a heroes grid config layout.

## Usage

The CLI supports `--help` option and each subcommand supports it.
There're several main command:
* `heroes-grid` which updates config files
* `startup` which has subcommands to control autorun of the program on OS startup (currently Windows only)


## Example

Run the application, which starts and periodically (each hour) generates new heroes grid in the provided order.

```
d2tool heroes-grid --periodic --positions 5,4,3,2,1
```

Register the application to run on startup and update heroes grid as described above.

```
d2tool startup register --periodic --positions 5,4,3,2,1
```

