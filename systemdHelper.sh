#!/bin/bash

if [[ $(id -u) -ne 0 ]]; then
    echo "[CRIT] You need to execute the script with root privileges"
    exit 1
fi

echo "Select the action you want to perform:\n\n  1. Create the systemd unit\n  2. Remove the systemd unit" 
read choice
case "$choice" in
"1")
    create_systemd_unit()
    ;;
"2")
    delete_systemd_unit()
    ;;
*)
    echo "wrong choice."
    exit 1
    ;;
esac

function create_systemd_unit {
    # Creates the configuration folder
    mkdir -p /etc/granti/

    # Creates the default configuration file if not present
    ls /etc/granti/granti.toml > /dev/null
    if [[ $? -ne 0]]; then
        cat << EOF > /etc/granti/granti.toml
        Logfile = "/var/log/granti.log"
        DatabasePath = "/var/granti.db"
        LogExistTimeout = "1s"

        [[jail]]
        Name = "myjail"
        Enabled = true
        Regex = "^(?P<IP>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)) (?P<timestamp>\\d{10}) .*$"
        RegexBlacklist = ["curl","zmap"]
        BlacklistBanCommand = "/home/user/banBotIP.sh <IP>"
        IPWhitelist = ["127.0.0.1/8", "10.0.0.0/8", "172.16.0.0/16", "192.168.0.0/16"]
        IPGroupName = "IP"
        TsGroupName = "timestamp"
        LogFile = "/var/log/log.log"
        CounterMaxValue = 100
        FindTime = "60m"
        Burst = 0
        TsLayout = "Mon Jan 2 15:04:05 -0700 MST 2006"
        BanAction = "/home/user/banip.sh <IP>"
EOF
    fi

    # Creates the Granti systemd unit if not present
    ls /etc/systemd/system/granti.unit > /dev/null
    if [[ $? -ne 0]]; then
        cat << EOF > /etc/systemd/system/granti.unit
        [Unit]
        Description=Granti Service
        After=network.target syslog.service
        # Requires=syslog.service

        [Service]
        Type=notify
        ExecStart=/usr/local/bin/granti -c /etc/granti/granti.toml
        KillMode=process
        Restart=on-failure

        [Install]
        WantedBy=multi-user.target
        Alias=granti.service
EOF
    systemctl daemon-reload
    fi
}

function delete_systemd_unit {
    rm -f /etc/systemd/system/granti.unit
    systemctl daemon-reload
}