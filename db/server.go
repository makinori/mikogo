package db

type Server struct {
	_        struct{} `cbor:",toarray"`
	Address  string
	Channels []string
}

var Servers = cborCrud[Server]{
	bucket: "servers",
}
