# transmon

[![Build Status](https://travis-ci.org/albertrdixon/transmon.svg?branch=master)](https://travis-ci.org/albertrdixon/transmon)

Adapted from [pia_transmission_monitor](https://github.com/firecat53/pia_transmission_monitor) and written in go.

Transmon is just a simple program to run Transmission and OpenVPN using Private Internet Access. Transmon will make sure Transmission is bound to the OpenVPN tunnel and will update the peer port hourly with PIA's port forwarding api. Also provides an optional torrent cleaner that monitors and removes stalled and finished torrents.

```
usage: transmon [<flags>]

Keep your transmission ports clear!

Flags:
  --help         Show context-sensitive help (also try --help-long and --help-man).
  -C, --config=/etc/transmon/config.yml
                 config file
  -c, --cleaner  enable transmission cleaner thread
  -l, --log-level=info
                 log level. One of: fatal, error, warn, info, debug
```