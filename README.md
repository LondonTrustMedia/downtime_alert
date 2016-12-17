# Downtime Alert

This checks a bunch of services to ensure they're up. If they go down and stay down, it reports those issues via email, sms, etc.

Configuration is done through the `config.yaml` file. Take a look at `config.example.yaml`.

Written by Daniel <doaks@londontrustmedia.com>, so yell at me if you need any help with this.


## Checking Method

So let's go into some more detail about how we check whether we should alert for a service.

Note that this app is designed to be cron'd for once every minute.

### Webpages

1. First launch of the monitor.
    1. Detect page failure.
    2. Wait 10 seconds (default).
    3. Check if the page is still down. If it's down, mark it as a failure.
2. Second launch of the monitor.
    1. Detect page failure. Assume webpage is down and start alerting.

After this, the monitor will send three (default) alerts, one minute (default) apart, and then only send alerts once per 20 minutes until the issue is resolved.

### SOCKS Proxy and VPN Gateways

1. First launch of the monitor.
    1. Detect proxy/VPN failure.
2. Second launch of the monitor.
    1. Detect proxy/VPN failure.
3. Third launch of the monitor.
    1. Detect proxy/VPN failure. Assume service is down and start alerting.
