#!/bin/bash
# start-test-env.sh - Mock NMOS Test Ortamını Başlat

cd "$(dirname "$0")"
echo "🚀 Mock NMOS test ortamı başlatılıyor..."

docker-compose up -d

echo "⏳ Servislerin hazır olması bekleniyor..."
sleep 5

echo "✅ Servis sağlığı kontrol ediliyor..."
curl -s http://localhost:8080/health > /dev/null && echo "✓ Mock Node 1: OK" || echo "✗ Mock Node 1: FAILED"
curl -s http://localhost:8081/health > /dev/null && echo "✓ Mock Node 2: OK" || echo "✗ Mock Node 2: FAILED"
curl -s http://localhost:8082/health > /dev/null && echo "✓ Mock Registry: OK" || echo "✗ Mock Registry: FAILED"

echo ""
echo "🎉 Test ortamı hazır!"
echo "Mock Node 1: http://localhost:8080"
echo "Mock Node 2: http://localhost:8081"
echo "Mock Registry: http://localhost:8082"
echo ""
echo "Servisleri durdurmak için: docker-compose down"
