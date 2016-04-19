package main

const (
	reqTopic       = "verify-requests"
	resTopicBase   = "verify-result-%06d"
	workerChannel  = "workers"
	managerChannel = "manager"

	nsqdAddr       = "127.0.0.1:4150"
	nsqLookupAddr  = "127.0.0.1:4161" // fixme: could me more than 1, for HA
	nsqMaxInFlight = 100              // What exactly does this do? =)
)

// VerfifyReq is the type used for when a verfification request goes through NSQ
type VerfifyReq struct {
	Email       string `json:"email"`
	ResultTopic string `json:"result-topic"` // The NSQ topic we should publish the result to
}

// VerfifyRes is the type used for when a verfification results goes through NSQ
type VerfifyRes struct {
	Email     string `json:"email"`
	AddressOK bool   `json:"address-ok"`
	SMTPMsg   string `json:"smtp-msg"`
	Error     error  `json:"error"`
}
