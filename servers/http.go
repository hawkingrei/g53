package servers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/hawkingrei/G53/utils"
	"net"
	"net/http"
	"regexp"
)

// HTTPServer represents the http endpoint
type HTTPServer struct {
	config *utils.Config
	list   ServiceListProvider
	server *http.Server
}

// NewHTTPServer create a new http endpoint
func NewHTTPServer(c *utils.Config, list ServiceListProvider) *HTTPServer {
	s := &HTTPServer{
		config: c,
		list:   list,
	}

	router := mux.NewRouter()
	router.HandleFunc("/services", s.getServices).Methods("GET")
	router.HandleFunc("/services/{id}", s.getService).Methods("GET")
	router.HandleFunc("/services", s.addService).Methods("PUT")
	router.HandleFunc("/services/{id}", s.updateService).Methods("PATCH")
	router.HandleFunc("/services/{id}", s.removeService).Methods("DELETE")
	router.HandleFunc("/set/ttl", s.setTTL).Methods("PUT")

	s.server = &http.Server{Addr: c.HttpAddr, Handler: router}

	return s
}

// Start starts the http endpoint
func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) getServices(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(s.list.GetAllServices())
}

func (s *HTTPServer) getService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id, _ := vars["id"]

	service, err := s.list.GetService(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(service)
}

func (s *HTTPServer) addService(w http.ResponseWriter, req *http.Request) {
	service := NewService()
	if err := json.NewDecoder(req.Body).Decode(&service); err != nil {
		logger.Errorf("JSON decoding error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.validation(service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.list.AddService(service.Aliases, *service)
}

func (s *HTTPServer) removeService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	id, _ := vars["id"]

	if err := s.list.RemoveService(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

}

func (s *HTTPServer) updateService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	id, _ := vars["id"]

	service, err := s.list.GetService(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(req.Body).Decode(&service); err != nil {
		logger.Errorf("JSON decoding error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.validation(&service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// todo: this probably needs to be moved. consider stop event in the
	// middle of sending PATCH. container would not be removed.
	s.list.AddService(id, service)

}

func (s *HTTPServer) setTTL(w http.ResponseWriter, req *http.Request) {
	var value int
	if err := json.NewDecoder(req.Body).Decode(&value); err != nil {
		logger.Errorf("Decoding error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.config.Ttl = value

}

func (s *HTTPServer) validation(service *Service) error {
	err := validateDomainType(service)
	if err != nil {
		return err
	}
	err = validateDomainValue(service)
	if err != nil {
		return err
	}
	return nil
}
func validateDomainType(service *Service) error {
	switch service.RecordType {
	case "A":
		if net.ParseIP(service.Value) == nil {
			logger.Debugf("Property \"Value\" is NOT IP")
			return errors.New("Property \"Value\" is NOT IP")
		}
	case "CNAME":
		if !validateDomainName(service.Value) {
			return errors.New("Property \"Value\" is wrong")
		}
	default:
		return errors.New("Property \"Record type\" is required or wrong")
	}
	return nil
}

func validateDomainValue(service *Service) error {
	if service.Aliases == "" {
		logger.Debugf("Property \"Aliases\" is required")
		return errors.New("Property \"Aliases\" is required")
	}
	if service.TTL <= 0 {
		logger.Debugf("Property \"TTL\" is required")
		return errors.New("Property \"TTL\" is required")
	}
	return nil
}

func validateDomainName(domain string) bool {
	RegExp := regexp.MustCompile(`^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z
 ]{2,3})$`)
	return RegExp.MatchString(domain)
}
