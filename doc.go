// Transmon is just a simple program to run Transmission and OpenVPN using Private Internet Access.
// Transmon will make sure Transmission is bound to the OpenVPN tunnel and will update the peer port
// hourly with PIA's port forwarding api. Also provides an optional torrent cleaner that monitors and
// removes stalled and finished torrents.
//
//      usage: transmon [<flags>]
//
//      Keep your transmission ports clear!
//
//      Flags:
//        --help         Show context-sensitive help (also try --help-long and --help-man).
//        -C, --config=/etc/transmon/config.yml
//                       config file
//        -l, --log-level=info
//                       log level. One of: fatal, error, warn, info, debug
//
package main

const version = "v0.2.2"
