package server

import (
	"context"
	api "github.com/tPhume/proglog/api/v1"
	"google.golang.org/grpc"
)

var _ api.LogServer = (*grpcServer)(nil)

type Config struct {
	CommitLog CommitLog
}

type CommitLog interface {
	Append(record *api.Record) (uint64, error)
	Read(offset uint64) (*api.Record, error)
}

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	grsv := grpc.NewServer(opts...)

	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}

	api.RegisterLogServer(grsv, srv)
	return grsv, nil
}

type grpcServer struct {
	*Config
}

func newgrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config: config,
	}

	return srv, nil
}

func (s *grpcServer) Produce(ctx context.Context, request *api.ProduceRequest) (*api.ProduceResponse, error) {
	offset, err := s.CommitLog.Append(request.Record)
	if err != nil {
		return nil, err
	}

	return &api.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) ProduceStream(server api.Log_ProduceStreamServer) error {
	for {
		req, err := server.Recv()
		if err != nil {
			return err
		}

		res, err := s.Produce(server.Context(), req)
		if err != nil {
			return err
		}

		if err = server.Send(res); err != nil {
			return err
		}
	}
}

func (s *grpcServer) Consume(ctx context.Context, request *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	record, err := s.CommitLog.Read(request.Offset)
	if err != nil {
		return nil, err
	}

	return &api.ConsumeResponse{Record: record}, nil
}

func (s *grpcServer) ConsumeStream(request *api.ConsumeRequest, server api.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-server.Context().Done():
			return nil
		default:
			res, err := s.Consume(server.Context(), request)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}

			if err = server.Send(res); err != nil {
				return err
			}

			request.Offset++
		}
	}
}
