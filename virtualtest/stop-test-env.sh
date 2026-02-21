#!/bin/bash
# stop-test-env.sh - Mock NMOS Test Ortamını Durdur

cd "$(dirname "$0")"
echo "🛑 Mock NMOS test ortamı durduruluyor..."

docker-compose down

echo "✅ Tüm servisler durduruldu."
