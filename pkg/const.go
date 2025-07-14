package pkg

const (
	STORE_TYPE_FILE          = "file"
	STORE_TYPE_CM            = "configmap"
	STORE_TYPE_MYSQL         = "mysql"
	PROBE_FUNC_DNS           = "dns"
	PROBE_FUNC_TCP           = "tcp"
	PROBE_FUNC_HTTP          = "http"
	PROBE_FUNC_PING          = "ping"
	PROBE_FUNC_UDP           = "udp"
	RPOBE_RUN_TYPE_HOST      = "host"
	RPOBE_RUN_TYPE_CONTAINER = "container"
)

const (
	ProbeSuccess   float64 = 1
	ProbeFailed    float64 = -1
	ErrReasonEmpty string  = "empty"
	ErrReasonTW    string  = "time_out"
)
