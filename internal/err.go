package server

import (
	"log"
)

// go routine to handle the error
func (s *Server) ErrorHandler(errorChan <-chan Error) {
	for err := range errorChan {
		log.Printf("[%s] Error received: %s", err.Conn, err.Msg)
		errMsg := EncodeError(err.Msg)
		err.Conn.Write(errMsg)
	}
}
