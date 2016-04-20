# go-email-address-check

https://goreportcard.com/report/github.com/oyvindsk/go-email-address-check

Try to find out if email addresses are valid, in a distributed (more than 1 ip) and throttled way

GO + NSQ

TODO:
 - Fix the topic creation delay
 - Submit now, get results later REST API
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


## NSQ
The api-and-manager takes rest calls with >0 email addresses to verify. This is a "job".
The api-and-manager then takes each
 - Email Verify Request: One (api-and-manager) to many (workers)
 - Email Verify Result: Many (workers) to one (api-and-manager)

### Response "problem"
The responses (results of the worker jobs) also need to get back to the manager.
NSQ topics are one-ways "streams" of messages so we are responsible for making sure the results get back to the manager.

There are several not-so-good ways to do this:
- Use one topic for each "job" (REST API call)
    - Let the manager decide the topic and send it with the message
    - makes it easy to write the manager code, 1 topic == 1 go routine
    - but discovery is slow, so it takes a while before the first response gets back
    - configurable? How low can we go?
- Skip nsqlookup and embed the nsqd address in the message
    - Seems kind of hacky
    - Can't then publish to nsqd on localhost??
- Send all responses to 1 known topic
    - More work then to code the manager, since it has to send the messages to the right go routines
    - Slower? Does it matter??
    - Could also let every go routine get a copy and filter out those intended for it
        - Easier? A lot of message duplication
        - Or maybe ask each go routine in order: "is this msg headed for you?"



### NSQ Overview

                                                              +---+
                                                              |   |
                                                              | N |
                                                              | S |
    +--------------------------------------------+            | Q |
    |                                            |            |   |            +-----------+
    |                                            | ------------------------->  |           |
    |                                            |            |   |            |   Worker  |
    |    API ENDPOINT & MANAGER                  | <------------------------+  |           |
    |                                            |            |   |            +-----------+
    |                                            |            |   |
    |    Takes API requests, aka jobs            |            |   |            +-----------+
    |    and sends Email Verification Requests   | +------------------------>  |           |
    |    to the Workers through NSQ              |            |   |            |   Worker  |
    |                                            | <------------------------+  |           |
    |                                            |            |   |            +-----------+
    +--------------------------------------------+            |   |
                                                              |   |
                                                              |   |
                                                              |   |
                                                              |   |
                                                              |   |
                                                              +---+
