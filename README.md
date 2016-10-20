# Service Monitor

This program checks a bunch of services for issues, then reports those issues via email, sms, etc.

Configuration is done through the `config.yaml` file, and an example of one can be seen at `config.example.yaml`.


## TODO

* OpenVPN monitor
* L2TP monitor
* PPTP monitor

* Allow directing from one config file to another (so instead of having auth details spread across 16 files you could have the auth in one file, and then point towards that file from a bunch of others)

* Investigate why the SOCKS proxy gives me the number of username/password failures that it does, and scale back the checking as appropriate. Once every five minutes maybe?