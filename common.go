package main

const (
	reqTopic       = "verify_requests"
	resTopicBase   = "verify_result-%06d"
	workerChannel  = "email_adress_verification"
	managerChannel = "results_gathering"

	nsqdAddr       = "127.0.0.1:4150"
	nsqLookupAddr  = "127.0.0.1:4161" // fixme: could me more than 1, for HA
	nsqMaxInFlight = 100              // What exactly does this do? =)
)

// VerifyReq is the type used for when a verfification request goes through NSQ
type VerifyReq struct {
	Email       string `json:"email"`
	ResultTopic string `json:"result-topic"` // The NSQ topic we should publish the result to
}

// VerifyRes is the type used for when a verfification results goes through NSQ
type VerifyRes struct {
	Email     string `json:"email"`
	AddressOK bool   `json:"address-ok"`
	SMTPMsg   string `json:"smtp-msg"`
	Error     string  `json:"error"` // Can it not be of type error ??
}
