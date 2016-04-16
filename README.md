# go-email-address-check

Try to find out if email addresses are valid, in a distributed (more than 1 ip) and throttled way

GO + NSQ

    +----------------------------+           +-----------------------------+
    |                            |           |                             |
    |                            |           |                             |
    |   API Server               |           |   Email Validator           |
    |   NSQ Producer / Consumer  +-+         |   NSQ Producer / Consumer   |
    |                            | |         |                             |
    |                            | |         |                             |
    +----------------------------+ |         +-------+---------------------+
                                   |                 |
                                   |                 |
                                   |                 |
                                   |                 |
    +---------------------+        |       +---------+----------------------+
    |                     |        +-------+                                |
    |                     |                |                                |
    |    NSQ Daemon /     |                |   DNS Lookup service / Cache   |
    |    Lookup service   |                |   NSQ Producer / Consumer      |
    |       etc ..        +----------------+                                |
    |                     |                |                                |
    +---------------------+                +--------------------------------+
