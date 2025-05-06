package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"pills-taking-reminder/internal/api/grpc/pb"
	"pills-taking-reminder/internal/usecase"
	"pills-taking-reminder/pkg/mw"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	pb.UnimplementedPTRServiceServer
	scheduleUseCase *usecase.ScheduleUseCase
	logger          *slog.Logger
	server          *grpc.Server
}

func NewGRPCServer(useCase *usecase.ScheduleUseCase, logger *slog.Logger) *GRPCServer {
	return &GRPCServer{
		scheduleUseCase: useCase,
		logger:          logger,
	}
}

func (s *GRPCServer) CreateSchedule(ctx context.Context, req *pb.ScheduleRequest) (*pb.ScheduleIDResponse, error) {
	s.logger.Info("got schedule creating request in grpc",
		slog.String("medicine", req.MedicineName),
		slog.Int("frequency", int(req.Frequency)),
		slog.Int64("user_id", req.UserId))

	input := usecase.ScheduleInput{
		MedicineName: req.MedicineName,
		Frequency:    int(req.Frequency),
		Duration:     int(req.Duration),
		UserID:       req.UserId,
	}

	id, err := s.scheduleUseCase.CreateSchedule(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			s.logger.Debug("schedule creation request rejected in gRPC", slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, "Invalid input parameters")
		case errors.Is(err, usecase.ErrScheduleExists):
			s.logger.Debug("schedule creation request rejected in gRPC", slog.String("error", err.Error()))
			return nil, status.Error(codes.AlreadyExists, "Schedule already exists")
		default:
			s.logger.Error("failed to create schedule in gRPC", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}
	return &pb.ScheduleIDResponse{
		ScheduleId: id,
	}, nil
}

func (s *GRPCServer) GetSchedulesIDs(ctx context.Context, req *pb.UserIDRequest) (*pb.ScheduleIDList, error) {
	s.logger.Info("got getting schedule ids request in grpc",
		slog.Int64("user_id", req.UserId))

	scheduleIDs, err := s.scheduleUseCase.GetScheduleIDs(ctx, req.UserId)

	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			s.logger.Debug("request for getting schedule IDs rejected in gRPC", slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, "Invalid input parameters")
		default:
			s.logger.Error("failed to get schedule IDs in gRPC", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}
	return &pb.ScheduleIDList{
		ScheduleIds: scheduleIDs,
	}, nil
}

func (s *GRPCServer) GetNextTakings(ctx context.Context, req *pb.UserIDRequest) (*pb.TakingList, error) {
	s.logger.Info("got GetNextTakings request in groc",
		slog.Int64("user_id", req.UserId))

	takings, err := s.scheduleUseCase.GetNextTakings(ctx, req.UserId)

	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			s.logger.Debug("request for getting next takings rejected in grpc", slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, "Invalid input parameters")
		default:
			s.logger.Error("error getting next takings in grpc", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}

	pbTakings := make([]*pb.Taking, len(takings))
	for i, taking := range takings {
		pbTakings[i] = &pb.Taking{
			MedicineName: taking.MedicineName,
			TakingTime:   taking.TakingTime,
		}
	}

	return &pb.TakingList{
		Takings: pbTakings,
	}, nil

}

func (s *GRPCServer) GetSchedule(ctx context.Context, req *pb.ScheduleIDRequest) (*pb.ScheduleResponse, error) {
	s.logger.Info("got GetSchedule request in grpc",
		slog.Int64("user_id", req.UserId),
		slog.Int64("schedule_id", req.ScheduleId))

	schedule, err := s.scheduleUseCase.GetSchedule(ctx, req.UserId, req.ScheduleId)

	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrScheduleNotFound):
			s.logger.Debug("request for getting schedule rejected in gRPC", slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, "Schedule was not found")
		case errors.Is(err, usecase.ErrInvalidInput):
			s.logger.Debug("request for getting schedule rejected in gRPC", slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, "Invalid input parameters")
		default:
			s.logger.Error("failed to get schedule in gRPC", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}

	return &pb.ScheduleResponse{
		Id:           schedule.ID,
		MedicineName: schedule.MedicineName,
		StartDate:    schedule.StartDate,
		EndDate:      schedule.EndDate,
		UserId:       schedule.UserID,
		TakingTime:   schedule.TakingTimes,
	}, nil
}

func (s *GRPCServer) Run(addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(mw.UnaryServerInterceptor(s.logger)),
		grpc.StreamInterceptor(mw.StreamServerInterceptor(s.logger)))

	pb.RegisterPTRServiceServer(s.server, s)

	reflection.Register(s.server)

	s.logger.Info("grpc server started", slog.String("address", addr))

	return s.server.Serve(listen)
}

func (s *GRPCServer) Stop() {
	if s.server != nil {
		s.logger.Info("stopping grpc server...")
		s.server.GracefulStop()
	}
}
