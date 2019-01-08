# event-framework-subscriber
KW Event Framework Subscriber written in GO, connect your application through unix socket, and send message once connected to subscribe.

Subscribe message format:
```
subscribe:topic-name;
```

To disconnect send message:
```
quit!;
```

Install:
- Download your OS latest release binary from releases page here: https://github.com/KWRI/event-framework-subscriber/releases
```
# example linux amd64
curl -L -o /bin/event-framework-subscriber https://github.com/KWRI/event-framework-subscriber/releases/download/v1.0.0/event-framework-subscriber_linux_amd64

# move file to /usr/bin
sudo mv ./event-framework-subscriber /usr/bin/event-framework-subscriber

# add permission
sudo chmod +x /usr/bin/event-framework-subscriber
```


Run:
```
GOOGLE_APPLICATION_CREDENTIALS=/Path_TO_GOOGLE_CREDENTIALS_FILE/google-secret.json event-framework-subscriber -project=event-framework -port=tcp_port_number_default_9001
```

Alternatively you can export GOOGLE_APPLICATION_CREDENTIALS to your OS environment variable from bash profile or using below script before run event-framework-subscriber
```
export GOOGLE_APPLICATION_CREDENTIALS=PATH_TO_GOOGLE_SECRET_CREDENTIAL_JSON_FILE/google-secret.json
```


