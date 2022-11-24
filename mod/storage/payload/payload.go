package payload

type RedisComm struct {
	Timestamp string
	Type      string
	SubType   string
}

type RedisNetPayload struct {
	RedisComm
	SendIP    string
	RecvIP    string
	Protocol  string
	TimeDelta float64
}

type RedisStructHostPayload struct {
	RedisComm
	OperatingSystem string
	OSType          string
	Architecture    string
	Name            string
	NCPU            int
	MemTotal        int64
	KernelVersion   string
}

type RedisStructNetPayload struct {
	RedisComm
	ID   string
	Name string
}

type RedisStructContNetPortPayload struct {
	PrivatePort uint16
	PublicPort  uint16
	Type        string
}

type RedisStructContPayload struct {
	RedisComm
	ID          string
	Name        string
	Image       string
	CPUPerc     string
	Locale      string
	Timezone    string
	IPAddresses map[string]string
	Ports       map[string][]RedisStructContNetPortPayload
}
