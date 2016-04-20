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

## NSQ
### Response "problem"
The responses (results of the worker jobs) also need to get back to the manager.
So either:
 - Use on topic for each "job" (REST API call)
  - Let the manager decide the topic and send it with the message
  - makes it easy to write the manager code, 1 topic == 1 go routine
  - but discovery is slow, so it takes a while before the first response gets back
  - configurable? How low can we go?
- Skip nsqlookup and embed the nsqd address in the message
 - Seems kind of hacky
 - Can't publish then to nsqd on localhost??
- Send all responses to 1 known topic
 - More work then to code the manager, since it has to send the messages to the right go routines
 - Slower? Does it matter??
 - Could also let every go routine get a copy and filter out those intendet for it
  - Easier? A lot of message duplication
 - Or maybe ask each go routine in order: "is this msg headed for you?"


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
