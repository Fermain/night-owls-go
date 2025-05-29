#!/bin/bash
# Script to check memory and resource usage on the droplet

echo "=== Memory Information ==="
free -h

echo -e "\n=== Memory Usage Details ==="
cat /proc/meminfo | grep -E "MemTotal|MemFree|MemAvailable|SwapTotal|SwapFree"

echo -e "\n=== Current Memory Usage by Process ==="
ps aux --sort=-%mem | head -10

echo -e "\n=== Docker Memory Usage ==="
docker system df

echo -e "\n=== Docker Container Stats ==="
docker stats --no-stream --all

echo -e "\n=== Disk Space ==="
df -h 