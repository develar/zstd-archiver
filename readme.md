```
usage: zstd-archiver [<flags>] <command> [<args> ...]

Compress and decompress directory using tar and zstd

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Commands:
  help [<command>...]
    Show help.


  compress --source=SOURCE --output=OUTPUT [<flags>]
    compress directory to tar.zst file

    -s, --source=SOURCE       source directory
    -o, --output=OUTPUT       output file
    -l, --level=19            compression level (1-19)
        --block-size=32MB     solid block size
        --system-independent  whether to ignore time, xattr and other system
                              dependent info

  decompress --input=INPUT --output=OUTPUT
    decompress tar.zst file

    -i, --input=INPUT    input file
    -o, --output=OUTPUT  output directory
```