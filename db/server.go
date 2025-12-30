package db

type Server struct {
	_        struct{} `cbor:",toarray"`
	Address  string
	Channels []string
}

var Servers = cborCrud[Server]{
	bucket: "servers",
}

func GetServerByAddress(address string) (string, Server, error, bool) {
	servers, err := Servers.GetAll()
	if err != nil {
		return "", Server{}, err, false
	}

	for name, server := range servers.AllFromBack() {
		if server.Address == address {
			return name, server, nil, true
		}
	}

	return "", Server{}, nil, false
}
