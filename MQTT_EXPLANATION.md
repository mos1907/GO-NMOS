# MQTT Nedir ve Neden Önemli?

## 🎯 MQTT Ne İşe Yarar?

MQTT (Message Queuing Telemetry Transport), **hafif bir mesajlaşma protokolü**. Bu projede **realtime event notification** için kullanılıyor.

## 📡 Bu Projede MQTT Nasıl Çalışıyor?

### Backend Tarafı

Flow'lar değiştiğinde (create/update/delete), backend otomatik olarak MQTT broker'a event gönderir:

```
go-nmos/flows/events/all          → Tüm flow event'leri
go-nmos/flows/events/flow/{id}    → Belirli bir flow'un event'leri
```

**Event Formatı:**
```json
{
  "event": "created|updated|deleted",
  "flow_id": "uuid-here",
  "flow": { /* flow data */ },
  "diff": { /* değişen alanlar */ },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Frontend Tarafı

Frontend, MQTT WebSocket üzerinden bu event'leri dinler ve **UI'yı otomatik günceller**:

- ✅ Flow listesi anında yenilenir
- ✅ Dashboard summary güncellenir
- ✅ Başka kullanıcıların yaptığı değişiklikler görünür

## 🔥 MQTT Aktif Olması Neden Önemli?

### 1. **Realtime Collaboration** (Çoklu Kullanıcı Desteği)
```
Kullanıcı A: Flow'u günceller
    ↓
MQTT event gönderilir
    ↓
Kullanıcı B: UI'sında otomatik güncellenir (sayfa yenilemeden!)
```

**MQTT YOKSA:**
- Her kullanıcı manuel "Refresh" butonuna basmalı
- Değişiklikler anında görünmez
- Çoklu kullanıcı senaryosunda sorunlu

**MQTT VARSA:**
- Değişiklikler anında tüm kullanıcılara yansır
- Sayfa yenileme gerekmez
- Gerçek zamanlı işbirliği mümkün

### 2. **External System Integration** (Dış Sistem Entegrasyonu)

Başka sistemler de MQTT'ye subscribe ederek flow değişikliklerini dinleyebilir:

```python
# Örnek: Python script flow değişikliklerini dinliyor
import paho.mqtt.client as mqtt

def on_message(client, userdata, msg):
    event = json.loads(msg.payload)
    if event['event'] == 'created':
        # Yeni flow oluşturuldu, başka bir sisteme bildir
        notify_external_system(event['flow'])

client = mqtt.Client()
client.connect("mqtt-broker", 1883)
client.subscribe("go-nmos/flows/events/all")
client.on_message = on_message
client.loop_forever()
```

**Kullanım Senaryoları:**
- Monitoring sistemleri flow değişikliklerini loglar
- Automation script'leri flow oluşturulunca tetiklenir
- BCC sistemleri flow güncellemelerini alır
- Alert sistemleri collision'ları izler

### 3. **Performance** (Performans)

**MQTT YOKSA:**
- Frontend sürekli polling yapmalı (her 5-10 saniyede bir API çağrısı)
- Gereksiz network trafiği
- Server yükü artar

**MQTT VARSA:**
- Push-based: Sadece değişiklik olduğunda mesaj gönderilir
- Daha az network trafiği
- Daha iyi performans

### 4. **Offline Support** (Çevrimdışı Desteği)

MQTT broker mesajları queue'da tutabilir:
- Client bağlantısı kesilse bile
- Yeniden bağlandığında missed event'leri alabilir
- (Bu projede şu an implement edilmedi ama eklenebilir)

## ⚙️ MQTT Aktif Etmek

### Backend `.env` Dosyası:
```bash
MQTT_ENABLED=true
MQTT_BROKER_URL=tcp://mqtt:1883
MQTT_TOPIC_PREFIX=go-nmos/flows/events
```

### Docker Compose:
MQTT servisi zaten `docker-compose.yml`'de tanımlı:
```yaml
mqtt:
  image: eclipse-mosquitto:2
  ports:
    - "1883:1883"    # MQTT
    - "9001:9001"    # WebSocket (frontend için)
```

## 🎯 Sonuç

**MQTT aktif olması önemli çünkü:**

1. ✅ **Realtime updates** - UI anında güncellenir
2. ✅ **Multi-user support** - Çoklu kullanıcı senaryosu sorunsuz çalışır
3. ✅ **External integration** - Dış sistemler entegre edilebilir
4. ✅ **Better performance** - Polling yerine push-based
5. ✅ **Production-ready** - Gerçek dünya senaryoları için gerekli

**MQTT olmadan da çalışır ama:**
- ❌ Her kullanıcı manuel refresh yapmalı
- ❌ Çoklu kullanıcı senaryosunda sorunlu
- ❌ Dış sistem entegrasyonu zor
- ❌ Daha fazla server yükü

## 💡 Öneri

**Production ortamında MQTT'yi aktif etmenizi şiddetle tavsiye ederim.** Özellikle:
- Birden fazla kullanıcı varsa
- Dış sistemlerle entegrasyon planlanıyorsa
- Realtime updates önemliyse

MQTT opsiyonel ama **production için kritik bir özellik**.
