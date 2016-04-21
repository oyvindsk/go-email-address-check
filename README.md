# go-email-address-check

https://goreportcard.com/report/github.com/oyvindsk/go-email-address-check

Try to find out if email addresses are valid, in a distributed (more than 1 ip) and throttled way

GO + NSQ

TODO:
 - Worker: Check a bs address
 - Worker: Return the attachment size from EHLO? http://stackoverflow.com/questions/10006226/how-to-retrieve-fixed-upper-limit-on-message-size-smtp-server
 - Worker: Add a checkedAt timestamp
 - API: Submit now, get results later REST API
 - Worker: Throttling of SMTP connections: per machine and per domain per machine
 - DNS MX caching
 - DNS mailserver lookup caching?
 - ASCII nsq overview
 - Use nsqlookupd and more nsqd's (And then we also need to fix the topic creation delay, see further down)

## Running
### NSQ
We only use 1 nsqd pluss the nsq admin at the moment. This should probably change in the future to make it more
resilient and distributed. The standard NSQ setup is to run 1 nsqd with every producer
(api-server and verify-worker for this project), and then a few nsqlookupd's to facilitate discovery.
However since we run on docker it's a little difficult to force the nsqd's to run on the right machine, so we just run one instead and point everything to it.

    nsqd &
    nsqadmin &


### API Server
This is the REST API endpoint and the manager that takes care of splitting a "job" (rest api call) into several small
requests for the workers to process

    go run api-server.go common.go ADDRESS-OF-NSQD

### Worker

    go run verify-client.go common.go ADDRESS-OF-NSQD

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
 - Email Verify Request: One (api-server) to many (verify-workers)
 - Email Verify Result: Many (verify-workers) to one (api-server)

### Response "problem"
(Note: Right now we just use 1 nsqd, so this is not a problem)

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
