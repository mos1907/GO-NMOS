import mqtt from "mqtt";

let client = null;
let reconnectTimer = null;

export function connectMQTT(wsUrl, topicPrefix, onMessage) {
  if (client && client.connected) {
    return client;
  }

  if (client) {
    client.end();
  }

  try {
    client = mqtt.connect(wsUrl, {
      clientId: `go-nmos-frontend-${Date.now()}`,
      reconnectPeriod: 5000,
      connectTimeout: 10000,
    });

    client.on("connect", () => {
      console.log("MQTT connected");
      client.subscribe(`${topicPrefix}/all`, (err) => {
        if (err) console.error("MQTT subscribe error:", err);
      });
    });

    client.on("message", (topic, message) => {
      try {
        const event = JSON.parse(message.toString());
        if (onMessage) {
          onMessage(event);
        }
      } catch (e) {
        console.error("MQTT message parse error:", e);
      }
    });

    client.on("error", (err) => {
      console.error("MQTT error:", err);
    });

    client.on("close", () => {
      console.log("MQTT disconnected");
    });

    client.on("offline", () => {
      console.log("MQTT offline");
    });

    return client;
  } catch (e) {
    console.error("MQTT connection failed:", e);
    return null;
  }
}

export function disconnectMQTT() {
  if (client) {
    client.end();
    client = null;
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

export function isMQTTConnected() {
  return client && client.connected;
}
