# go-email-address-check

https://goreportcard.com/report/github.com/oyvindsk/go-email-address-check

Try to find out if email addresses are valid, in a distributed (more than 1 ip) and throttled way

GO + NSQ

TODO:
 - Throttling of SMTP connections: per machine and per domain per machine
 - Json messages (or?)
 - DNS MX caching
 - DNS mailserver lookup caching?
 - ASCII nsq overview


## Architecture overview

                                                         +----------+
                                                         |          |
                                         +---------------+  Worker  |        +-----------------------+
                                         |               |          +--------+                       |
    +-----------------+                  |               +----------+        | DNS lookup service    |
    |                 |                  |                                   | (Not implemented yet) |
    | API & "manager" +------------------+                                   |                       |
    |                 |                  |                                   +--------+--------------+
    +-----------------+                  |               +----------+                 |
                                         |               |          |                 |
                                         +---------------+ Worker   +-----------------+
                             (JSON over NSQ)             |          |
                                                         +----------+


                                                         +----------+
                                                         |          |
                                                         | Worker   |
                                                         |          |
                                                         +----------+
                                                         +----------+
                                                         |          |
                                                         | Worker   |
                                                         |          |
                                                         +----------+
