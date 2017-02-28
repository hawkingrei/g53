package utils

// Service represents a container and an attached DNS record
// service(recode_type: "A",value: []string{"127.0.0.1","127.0.0.1"},Aliases: "www.duitang.net" ))
type Service struct {
	RecordType string
	Value      string
	TTL        int
	Private    bool
	Aliases    string
}
