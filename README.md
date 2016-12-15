# Downtime Alert

This checks a bunch of services to ensure they're up. If they go down and stay down, it reports those issues via email, sms, etc.

Configuration is done through the `config.yaml` file. Take a look at `config.example.yaml`.

Written by Daniel <doaks@londontrustmedia.com>, so yell at me if you need any help with this.


## Databases and Data Stores

So, this program stores two separate types of data between instances:

1. Data that we can throw away with no issue. For example, which set of credentials did we last use for the SOCKS proxy?
2. Data which must/should be preserved. Mostly, historical instances of downtime, webpage speeds, etc. SLIs which we can use later to make decisions or view how our services are going overall.

These must be preserved across multiple instances of the monitor. As such, this presents a problem: How do we preserve both of these things across multiple instances, while keeping the monitor lightweight and basically to a single binary?

Simple! If the monitor detects that there's no existing 'data manager' running, it starts itself as one and stays available for other monitors to connect to it. Communication between different instances of the monitor is done through Go TCP RPC.


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
