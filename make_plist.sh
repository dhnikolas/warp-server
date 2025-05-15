#!/usr/bin/env bash

cat <<END >application.ru.server.warp.daemon.plist
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"\>
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>application.ru.server.warp.daemon</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/warp-server</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

    <key>StandardOutPath</key>
    <string>/tmp/warp-server.out</string>

    <key>StandardErrorPath</key>
    <string>/tmp/warp-server.err</string>

    <key>HardResourceLimits</key>
    <dict>
        <key>NumberOfFiles</key>
        <integer>100</integer>
        <key>FileSize</key>
        <integer>1048576</integer> <!-- 1MB -->
    </dict>

    <key>SoftResourceLimits</key>
    <dict>
        <key>NumberOfFiles</key>
        <integer>50</integer>
    </dict>

    <key>ProcessType</key>
    <string>Interactive</string>

    <key>AbandonProcessGroup</key>
    <true/>
</dict>
</plist>
END
