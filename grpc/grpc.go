package grpc

import (
	"context"
	"fmt"
	"github.com/NethermindEth/juno/db"
	"github.com/NethermindEth/juno/grpc/gen"
	"github.com/NethermindEth/juno/utils"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	port uint16
	srv  *grpc.Server
	db   db.DB
	log  utils.SimpleLogger
}

func NewServer(port uint16, db db.DB, log utils.SimpleLogger) *Server {
	srv := grpc.NewServer()

	return &Server{
		srv:  srv,
		db:   db,
		port: port,
		log:  log,
	}
}

func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		// todo Stop() vs GracefulStop()
		s.srv.Stop()
	}()

	gen.RegisterDBServer(s.srv, handlers{s.db})

	return s.srv.Serve(lis)
}