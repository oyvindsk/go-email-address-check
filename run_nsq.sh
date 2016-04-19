cd _nsq-files
konsole --nofork --geometry=1200x300-0+0    -e bash -c '../nsq-bin/nsqlookupd; sleep 5'                                     2>/dev/null &
konsole --nofork --geometry=1200x300-0+350  -e bash -c '../nsq-bin/nsqd --lookupd-tcp-address=127.0.0.1:4160; sleep 5'      2>/dev/null &
konsole --nofork --geometry=1200x300-0+690  -e bash -c '../nsq-bin/nsqadmin --lookupd-http-address=127.0.0.1:4161; sleep 5' 2>/dev/null &
