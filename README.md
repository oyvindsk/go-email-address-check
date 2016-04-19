# go-email-address-check

https://goreportcard.com/report/github.com/oyvindsk/go-email-address-check

Try to find out if email addresses are valid, in a distributed (more than 1 ip) and throttled way

GO + NSQ

TODO:
 - Do SMTP lookups in parallell - go ..
 - Fix the topic creation delay
 - Submit now, get resultds later REST API
 - Throttling of SMTP connections: per machine and per domain per machine
 - DNS MX caching
 - DNS mailserver lookup caching?
 - ASCII nsq overview


## Architecture overview

         |                                                 +----------+
         |  (REST + JSON)                                  |          |
         |                                 +---------------+  Worker  |        +-----------------------+
         |                                 |               |          +--------+                       |
    +-----------------+                    |               +----------+        | DNS lookup service    |
    |                 |                    |                                   |  (Not implemented)    |
    | API & "manager" +--------------------+                                   |                       |
    |                 |  (JSON over NSQ)   |                                   +--------+--------------+
    +-----------------+                    |               +----------+                 |
                                           |               |          |                 |
                                           +---------------+ Worker   +-----------------+
                                                           |          |
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
