### Home Assistant Webhook Automation

The following automation.yaml is an automaion that will trigger when a webhook is detected. The automation will then send a persistent notification to the user's device with the message details beign the data from the webhook.

> Set the `WEBHOOK_URL` in the .env file to the webhook URL that you will use to trigger the automation. 
>   - `WEBHOOK_URL=http://192.168.1.240:8123/api/webhook/alrtly_webhook` For triggering the automation from the Home Assistant Webhook.

```yaml
alias: Alrtly Webhook Automation
description: ""
trigger:
  - platform: webhook
    allowed_methods:
      - POST
      - PUT
    local_only: true
    webhook_id: "alrtly_webhook" # This is the webhook ID that you will use to trigger the automation (It is an usually an auto-generated string)
condition: []
action:
  - service: notify.persistent_notification
    metadata: {}
    data:
      title: Alrtly NOTICE
      message: >-
        "{{trigger.json.title}}: {{trigger.json.content}}
        [{{trigger.json.priority}}] ({{trigger.json.source}})"
mode: single
```

### Audio File Paths

#### Original Audio Files

- Canada AlertReady: `http://localhost:8000/audio/ca.mp3`
- Australia Standard Emergency Warning Signal (SEWS): `http://localhost:8000/audio/au.mp3`
- New Zealand "The Sting": `http://localhost:8000/audio/nz.wav`

#### Alexa Compatible Audio Files

- Canada AlertReady: `http://localhost:8000/audio/alexa/ca.mp3`
- Australia Standard Emergency Warning Signal (SEWS): `http://localhost:8000/audio/alexa/au.mp3`
- New Zealand "The Sting": `http://localhost:8000/audio/alexa/nz.mp3`