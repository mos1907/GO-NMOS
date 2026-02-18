# What is MQTT and Why is it Important?

## 🎯 What Does MQTT Do?

MQTT (Message Queuing Telemetry Transport) is a **lightweight messaging protocol**. In this project, it's used for **realtime event notification**.

## 📡 How Does MQTT Work in This Project?

### Backend Side

When flows change (create/update/delete), the backend automatically sends events to the MQTT broker:

```
go-nmos/flows/events/all          → All flow events
go-nmos/flows/events/flow/{id}    → Events for a specific flow
```

**Event Format:**
```json
{
  "event": "created|updated|deleted",
  "flow_id": "uuid-here",
  "flow": { /* flow data */ },
  "diff": { /* changed fields */ },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Frontend Side

The frontend listens to these events via MQTT WebSocket and **automatically updates the UI**:

- ✅ Flow list is instantly refreshed
- ✅ Dashboard summary is updated
- ✅ Changes made by other users are visible

## 🔥 Why is Enabling MQTT Important?

### 1. **Realtime Collaboration** (Multi-User Support)
```
User A: Updates a flow
    ↓
MQTT event is sent
    ↓
User B: UI automatically updates (without page refresh!)
```

**WITHOUT MQTT:**
- Every user must manually click the "Refresh" button
- Changes are not visible immediately
- Multi-user scenarios are problematic

**WITH MQTT:**
- Changes are instantly reflected to all users
- No page refresh needed
- Real-time collaboration is possible

### 2. **External System Integration**

Other systems can also subscribe to MQTT to listen for flow changes:

```python
# Example: Python script listening to flow changes
import paho.mqtt.client as mqtt

def on_message(client, userdata, msg):
    event = json.loads(msg.payload)
    if event['event'] == 'created':
        # New flow created, notify another system
        notify_external_system(event['flow'])

client = mqtt.Client()
client.connect("mqtt-broker", 1883)
client.subscribe("go-nmos/flows/events/all")
client.on_message = on_message
client.loop_forever()
```

**Use Cases:**
- Monitoring systems log flow changes
- Automation scripts are triggered when flows are created
- BCC systems receive flow updates
- Alert systems monitor collisions

### 3. **Performance**

**WITHOUT MQTT:**
- Frontend must continuously poll (API call every 5-10 seconds)
- Unnecessary network traffic
- Increased server load

**WITH MQTT:**
- Push-based: Messages are sent only when changes occur
- Less network traffic
- Better performance

### 4. **Offline Support**

MQTT broker can queue messages:
- Even if client connection is lost
- Can receive missed events when reconnected
- (Not currently implemented in this project but can be added)

## ⚙️ Enabling MQTT

### Backend `.env` File:
```bash
MQTT_ENABLED=true
MQTT_BROKER_URL=tcp://mqtt:1883
MQTT_TOPIC_PREFIX=go-nmos/flows/events
```

### Docker Compose:
MQTT service is already defined in `docker-compose.yml`:
```yaml
mqtt:
  image: eclipse-mosquitto:2
  ports:
    - "1883:1883"    # MQTT
    - "9001:9001"    # WebSocket (for frontend)
```

## 🎯 Conclusion

**MQTT is important to enable because:**

1. ✅ **Realtime updates** - UI updates instantly
2. ✅ **Multi-user support** - Multi-user scenarios work smoothly
3. ✅ **External integration** - External systems can be integrated
4. ✅ **Better performance** - Push-based instead of polling
5. ✅ **Production-ready** - Essential for real-world scenarios

**It works without MQTT but:**
- ❌ Every user must manually refresh
- ❌ Multi-user scenarios are problematic
- ❌ External system integration is difficult
- ❌ Higher server load

## 💡 Recommendation

**I strongly recommend enabling MQTT in production environments.** Especially if:
- There are multiple users
- External system integration is planned
- Realtime updates are important

MQTT is optional but **a critical feature for production**.
