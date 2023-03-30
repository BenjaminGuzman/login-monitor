#!/usr/bin/env bash

if ! cd /etc/audit/rules.d; then
    echo "Rules should be downloaded in /etc/audit/rules.d"
    echo -e "\033[93mRun this script with root privileges\033[0m"
    exit 1
fi

declare -A urls # associative array of urls
urls=(
    ["Inmutable loginuid "]="https://raw.githubusercontent.com/linux-audit/audit-userspace/master/rules/11-loginuid.rules"
    ["Installers"]="https://raw.githubusercontent.com/linux-audit/audit-userspace/master/rules/44-installers.rules"
    ["pci-dss v3.1"]="https://raw.githubusercontent.com/linux-audit/audit-userspace/master/rules/30-pci-dss-v31.rules"
    ["Inmutable configuration"]="https://raw.githubusercontent.com/linux-audit/audit-userspace/master/rules/99-finalize.rules"
)

for ruleName in "${!urls[@]}"; do
    echo -e "Downloading rule \033[92m$ruleName\033[0m..."
    curl -O "${urls[$ruleName]}" --progress-bar
done

echo "Done."

echo -e "\nNow is time to update rules configuration (inside /etc/audit/rules.d) as needed."
echo -e "When you're done, \033[97mreboot\033[0m to apply changes"
