package servers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net"
	"net/http"

	"github.com/hawkingrei/G53/utils"
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

	if err := json.NewEncoder(w).Encode(s.list.GetAllServices()); err != nil {
		logger.Errorf("Encoding error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *HTTPServer) getService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id, ok := vars["id"]
	if !ok {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	service, err := s.list.GetService(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(service); err != nil {
		logger.Errorf("Encoding error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	id, ok := vars["id"]
	if !ok {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	if err := s.list.RemoveService(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

}

func (s *HTTPServer) updateService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	id, ok := vars["id"]
	if !ok {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

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
	if service.RecordType == "" || service.RecordType != "A" && service.RecordType != "CNAME" {
		logger.Debugf("Property \"Record type\" is required or wrong")
		return errors.New("Property \"Record type\" is required or wrong")
	}
	if service.Value == "" {
		logger.Debugf("Property \"Value\" is required")
		return errors.New("Property \"Value\" is required")
	}
	if service.RecordType == "A" && net.ParseIP(service.Value) == nil {
		logger.Debugf("Property \"Value\" is NOT IP")
		return errors.New("Property \"Value\" is NOT IP")
	}
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
