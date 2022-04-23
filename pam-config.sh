#!/usr/bin/env bash

exepath="/root/.login-monitor/login-monitor"
configpath="/etc/pam.d/sshd"

inclusion_mode="optional"
args=""

next_is_config="false"
for arg in "$@" ; do
  case $arg in
    --required)
      inclusion_mode="required"
      ;;
    --optional)
      inclusion_mode="optional"
      ;;
    --help)
      echo "Script to add PAM rule to execute login-monitor script executable a successful login"
      echo "Usage: pam-config.sh [options...] args"
      echo ""
      echo "Options:"
      echo "--required: Tells PAM to disallow user login if login-monitor executable fails (use for production)"
      echo "--optional: Tells PAM to allow user to login even if login-monitor executable fails (use for testing)"
      echo "--config-path: Path to the file to be modified."
      echo "               Example of config files: /etc/pam.d/sshd, /etc/pam.d/common-auth."
      echo "               Default: $configpath"
      echo ""
      echo -e "args is a \033[97msingle\033[0m string of all the arguments to be provided to login-monitor when executed"
      exit 0
      ;;
    --config-path)
      next_is_config="true"
      ;;
    *)
      if [ "$next_is_config" == "true" ]; then
        configpath="$arg"
        next_is_config="false"
        continue
      fi

      args="$arg"
      ;;
  esac
done

if [[ ! -f "$configpath" ]]; then
  echo -e "\033[93mFile $configpath doesn't exist\033[0m"
  exit 1
fi

if [[ ! -w "$configpath" ]]; then
  echo -e "\033[93mRun this script with root privileges\033[0m"
  exit 1
fi

if ! more "$configpath" | grep -q "$exepath"; then
  backup="$configpath.bak"
  echo "Back up is saved in $backup"
  cp "$configpath" "$backup"

  pam_config="session $inclusion_mode pam_exec.so seteuid $exepath $args"
  comment="monitor successful logins"
  echo -e "\n# $comment\n$pam_config" | sudo tee --append "$configpath" > /dev/null 2>&1 &&\
  echo -e "\033[92mSuccessfully added PAM module\033[0m"
else
  echo -e "\033[93mThe rule may already be added. No modification was made\033[0m"
fi