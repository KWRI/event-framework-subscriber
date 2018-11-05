# event-framework-subscriber
KW Event Framework Subscriber written in GO


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
GOOGLE_APPLICATION_CREDENTIALS=PATH_TO_GOOGLE_SECRET_CREDENTIAL_JSON_FILE/google-secret.json event-framework-subscriber -
subscriptions="subscriptions_name separated_by_space" -project=google-pubsub-project-name
```

Alternatively you can export GOOGLE_APPLICATION_CREDENTIALS to your OS environment variable from bash profile or using below script before run event-framework-subscriber
```
export GOOGLE_APPLICATION_CREDENTIALS=PATH_TO_GOOGLE_SECRET_CREDENTIAL_JSON_FILE/google-secret.json
```


