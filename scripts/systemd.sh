#!/bin/bash -e
# This script manages installation using SystemD, if available, or SysV as a fall-back.
#
# Usage:
#  systemd.sh [ install | remove | status | log ]

cd $(dirname $0)/..
[[ -n $VERSION ]] || export VERSION=unset

function ensureSystemdIsPresent
{
  if [[ $( uname -s ) != "Linux" ]]; then
    echo "SystemD can only be used on Linux systems."
    exit 1
  fi

  id=$(uname -msn)

  if command -v systemctl >/dev/null; then
    echo -n "$id has "
    systemctl --version | head -1

  elif [ -d /etc/init.d ]; then
    echo "$id has SysV."

  else
    echo "Neither SystemD nor SysV is available in this operating system."
    lsb_release -a
    exit 1

  fi
}

# doInstallConfig installs the config file
function doInstallConfig
{
  if [ ! -r config.toml ]; then
    echo "Missing: config.toml"
    echo "Copy & modify config.dev.toml as needed."
    exit 1
  fi
  sudo mkdir -p /etc/athens
  sudo install -v -o root -g root -m 644 config.toml /etc/athens
}

# doInstallBinary copies the Athens binary to /usr/local/bin with the necessary settings.
function doInstallBinary
{
  [ -f athens ] || make athens

  if [ ! -x /usr/local/bin/athens -o athens -nt /usr/local/bin/athens ]; then

    [ -f /etc/systemd/system/athens.service ] && sudo systemctl stop athens
    [ -f /etc/init.d/athens ] && sudo /etc/init.d/athens stop

    sudo install -v -o root -g root athens /usr/local/bin
  fi

  # Give the athens binary the ability to bind to privileged ports (e.g. 80, 443) as a non-root user:
  command -v setcap >/dev/null && sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/athens
}

# doInstallSystemd sets up the SystemD service unit.
function doInstallSystemd
{
  local rootPath=$(sed -nr 's/(RootPath) = (".*")/\2/p' /etc/athens/config.toml | xargs)
  sed -i "/ReadWritePaths/ s|=.*|=$rootPath|" scripts/service/athens.service

  sudo install -v -o root -g root -m 644 scripts/service/athens.service /etc/systemd/system
  sudo systemctl daemon-reload
  sudo systemctl enable athens
  sudo systemctl start athens
}

# doInstall builds and installs the binary as a SystemD unit
function doInstall
{
  doInstallConfig
  doInstallBinary
  doInstallSystemd
}

# doRemove deletes the SystemD unit and cleans up log files etc
function doRemove
{
  if [ -f /etc/systemd/system/athens.service ]; then
    sudo systemctl stop athens
    sudo rm -f /etc/systemd/system/athens.service
    # Reset systemctl
    sudo systemctl daemon-reload
    echo "SystemD installation was removed."

  elif [ -f /etc/init.d/athens ]; then
    sudo service athens stop
    sudo update-rc.d athens remove
    sudo rm -f /etc/init.d/athens
    echo "SysV installation was removed."

  fi

  sudo rm -rf /etc/athens /etc/ssl/athens /usr/local/bin/athens /var/log/athens.log /var/www/.athens
}

# doStatus shows what is installed, if anything, and whether it is running
function doStatus
{
  if [ -x /usr/local/bin/athens ]; then
    echo "Athens is /usr/local/bin/athens"
    /usr/local/bin/athens --version
  else
    echo "Athens is absent (no /usr/local/bin/athens)."
    exit 0
  fi

  if [ -f /etc/systemd/system/athens.service ]; then
    echo
    echo "SystemD: /etc/systemd/system/athens.service exists."
    sudo systemctl status athens ||:

  elif [ -f /etc/init.d/athens ]; then
    echo
    echo "SysV: /etc/init.d/athens exists."
    sudo service athens status ||:

  else
    echo "Athens is not installed as a service."
  fi
}

# showLog shows the relevant lines in syslog
function showLog
{
  if [ -x /usr/local/bin/athens ]; then
    echo "Athens is /usr/local/bin/athens"
    /usr/local/bin/athens --version
  else
    echo "Athens is absent (no /usr/local/bin/athens)."
    exit 0
  fi

  if [ -f /etc/systemd/system/athens.service ]; then
    fgrep athens /var/log/syslog | fgrep "$(date '+%b %d')"

  elif [ -f /etc/init.d/athens ]; then
    fgrep athens /var/log/syslog | fgrep "$(date '+%b %d')"

  else
    echo "Athens is not installed as a service."
  fi
}

### Main script ###doStatus

case $1 in
  install)
    ensureSystemdIsPresent; doInstall ;;

  remove|uninstall)
    ensureSystemdIsPresent; doRemove  ;;

  status)
    ensureSystemdIsPresent; doStatus  ;;

  log)
    ensureSystemdIsPresent; showLog  ;;

  *)
    echo "Usage: $0 [ install | remove | status | log"
    exit 1 ;;
esac
