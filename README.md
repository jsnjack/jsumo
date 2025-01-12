jsumo
=====

`jsumo` is a tool to quickly forward your logs from journalctl to SumoLogic. It uses
journalctl [cursor](https://www.freedesktop.org/software/systemd/man/latest/journalctl.html#-c)
to ensure that no logs are lost. If the access key for Sumologic REST API is provided,
it will automatically create a new source and a new collector for the logs based on the
hostname of the machine.

### Description
```
Usage:
  jsumo [flags]

Flags:
  -c, --category string            override source category with the given value
  -d, --debug                      enable debug mode
  -g, --grep string                pass grep pattern to journalctl command
  -h, --help                       help for jsumo
      --read-interval duration     interval to read logs from journalctl (default 5s)
      --upload-interval duration   interval to upload files to the receiver URL (default 2s)
  -r, --url string                 receiver URL. If empty, it will be fetched or created automatically using SumoLogic API
  -v, --version                    print version and exit
```

### Details
When `jsumo` is started, it will create a working directory in `~/.local/jsumo`.
This directory will contain the following files:
 - `jsumo-cursor`: This file will contain the cursor of the last log read from journalctl
 - `batch-*.zst.jsumo`: These files will contain the logs read from journalctl. The logs are compressed using zstd.

Sumologic recommends to limit the size of the uploaded logs to 1MB to avoid any
timeouts related to the log processing. When `jsumo` reads the logs from journalctl,
it will split the logs into multiple files based on the size of the logs.

When SIGINT is received, `jsumo` will attempt to gracefully shurdown. This means,
that:
 - if there is an active upload, it will wait for the upload to finish
 - if the log processing is active, it will wait for it to finish
 - there is a timeout which if reached, will force the shutdown

`jsumo` is designed to work with Sumologic HTTP Source, but it can be used with any
receiver URL that accepts POST requests with the logs in the body.

### Installation
 - Using [grm](https://github.com/jsnjack/grm)
    ```bash
    grm install jsnjack/jsumo
    ```
 - Download binary from [Release](https://github.com/jsnjack/jsumo/releases/latest/) page
