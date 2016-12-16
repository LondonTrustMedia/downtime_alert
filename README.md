# Downtime Alert

This checks a bunch of services to ensure they're up. If they go down and stay down, it reports those issues via email, sms, etc.

Configuration is done through the `config.yaml` file. Take a look at `config.example.yaml`.

Written by Daniel <doaks@londontrustmedia.com>, so yell at me if you need any help with this.


## Databases and Data Stores

We need to store data between different runs of the downtime alerter. Different runs of the downtime alerter can be active at one time, which make things very annoying and fidgety when using flat-file or similar file-based datastores.

How do we preserve info across multiple instances, while keeping the monitor lightweight, and make sure it doesn't depend on other stuff like heavy external databases?

Simple! If the monitor detects that there's no existing 'data manager' running, it starts itself as one and stays available for other monitors to connect to it. Communication between different instances of the monitor is done through Go TCP RPC.

### In-Memory vs Stored

Most data we have can simply be stored in-memory, such as the specific set of SOCKS creds we're using or where in the 'alert cycle' we are. We assume this can stay in-memory due to the presumably high reliability of the datastore itself.

Stored data are things that need to be preserved in-disk in some format. Historical information that lets us create graphs later, SLIs which we can use later to make decisions or view how our services are going overall. This is something I'm not approaching yet, but I'm looking into.


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
