package utils

import (
	"time"
)

// Service represents a container and an attached DNS record
// service(recode_type: "A",value: []string{"127.0.0.1","127.0.0.1"},Aliases: "www.duitang.net" ))
type Service struct {
	RecordType string
	Value      string
	TTL        int
	Aliases    string
}

type Entry struct {
	Aliases    string
	RecordType string
	Value      string
	TTL        int
	Time       time.Time
}

func EntryToServer(s *Entry) Service {
	return Service{(*s).RecordType, (*s).Value, (*s).TTL, (*s).Aliases}
}
