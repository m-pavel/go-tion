# Tion Breazer 3S Go API
There are thee backend available:
  - https://github.com/m-pavel/go-gattlib
  - https://github.com/paypal/gatt
  - https://github.com/muka/go-bluetooth <- Preferable
## Pairing
Device must be paired. E.g. using bluetoothctrl
## Reading state
```
    t = tionm.New("MAC")
    ts.t.Connect(timeout)
    defer ts.t.Disconnect()
    s, err := ts.t.ReadState(timeout)
    fmt.Println(s)
```
# Tion Breazer Home Assistant MQTT integration
## Sensors
```
- platform: mqtt
  name: "Temperature Inside (Tion)"
  state_topic: "nn/tion"
  value_template: "{{ value_json.temp_out }}"
  availability_topic: "nn/tion-aval"
  icon: "mdi:thermometer"
  unit_of_measurement: '°C'

- platform: mqtt
  name: "Temperature Outside (Tion)"
  state_topic: "nn/tion"
  value_template: "{{ value_json.temp_in }}"
  availability_topic: "nn/tion-aval"
  icon: "mdi:thermometer"
  unit_of_measurement: '°C'

- platform: mqtt
  name: "Temperature Target (Tion)"
  state_topic: "nn/tion"
  value_template: "{{ value_json.temp_target }}"
  availability_topic: "nn/tion-aval"
  icon: "mdi:thermometer"
  unit_of_measurement: '°C'

- platform: mqtt
  name: "Speed (Tion)"
  state_topic: "nn/tion"
  value_template: "{{ value_json.speed }}"
  availability_topic: "nn/tion-aval"
  icon: "mdi:fan"

```
## Control channel
Turn on/off
```
    {
      "payload_template": "{% if is_state('binary_sensor.tion' , 'off') %} \n  { \"on\": true }\n{% else %}\n  { \"on\": false }\n{% endif %}\n",
      "qos": 1,
      "topic": "nn/tion-control"
    }
```

