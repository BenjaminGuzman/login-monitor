#!/usr/bin/env bash

for arg in "$@" ; do
  case $arg in
    -h|--help)
      echo "Script to:"
      echo "1. Download and import public keys for recipients (provide URLs as positional arguments)"
      echo "2. Generate public key pair for this server so the sent emails are signed (--gen-key)"
      ;;
    --gen-key)
      echo "Generating input file..."
      host=$(hostname)
      passphrase=$(cat /proc/sys/kernel/random/uuid)
      passphrase_file="/root/$host-$RANDOM.txt"
      echo "$passphrase" > "$passphrase_file"
      if [[ -$? -ne 0 ]]; then
          echo -e "\033[93mRun this script with root privileges\033[0m"
          exit 1
      fi

      # https://www.gnupg.org/documentation/manuals/gnupg-devel/Unattended-GPG-key-generation.html
      echo -e "Key-Type: default\nSubkey-Type: default\nExpire-Date: 0\nPassphrase: $passphrase" >> /tmp/gen-key.txt
      echo -e "Name-Real: $host\nName-Email: $host@$host\n" >> /tmp/gen-key.txt
      echo "%commit" >> /tmp/gen-key.txt

      echo -e "Generating keys..."
      gpg --batch --gen-key /tmp/gen-key.txt
      rm -f /tmp/gen-key.txt
      echo Done.

      pub_out="/tmp/$host.pub.asc"
      gpg --export --armor "$host@$host" > "$pub_out"
      echo -e "Public key is saved to \033[97m$pub_out\033[0m and it is shown below for convenience."
      echo "(This is the one you import to your personal keyring to verify emails sent from this server)"
      more "$pub_out"
      echo -e "Passphrase is saved to \033[97m$passphrase_file\033[0m"
      echo -e "(This is the value for the key \"senderPassFile\" in the config file)"
      ;;
    *) # assume arg is a URL pointing to a public key
      echo -e "Downloading key \033[92m$arg\033[0m..."
      curl "$arg" --progress-bar | gpg --import
      ;;
  esac
done
