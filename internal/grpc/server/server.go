package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"pills-taking-reminder/internal/grpc/pb"
	"pills-taking-reminder/internal/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service interface {
	CreateSchedule(schedule models.ScheduleRequest) (int64, error)
	GetSchedulesIDs(userID int64) ([]int64, error)
	GetNextTakings(id int64) ([]models.Taking, error)
	GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error)
}

type GRPCServer struct {
	pb.UnimplementedPTRServiceServer
	service Service
	logger  *slog.Logger
	server  *grpc.Server
}

func NewGRPCServer(service Service, logger *slog.Logger) *GRPCServer {
	return &GRPCServer{
		service: service,
		logger:  logger,
	}
}

func (s *GRPCServer) CreateSchedule(ctx context.Context, req *pb.ScheduleRequest) (*pb.ScheduleIDResponse, error) {
	s.logger.Info("got schedule creating request in grpc",
		slog.String("medicine", req.MedicineName),
		slog.Int64("user_id", req.UserId))

	schedule := models.ScheduleRequest{
		MedicineName: req.MedicineName,
		Frequency:    int(req.Frequency),
		Duration:     int(req.Duration),
		UserID:       req.UserId,
	}

	id, err := s.service.CreateSchedule(schedule)
	if err != nil {
		s.logger.Error("error creating schedule in grpc", slog.String("error", err.Error()))
		return nil, err
	}
	return &pb.ScheduleIDResponse{
		ScheduleId: id,
	}, nil
}

func (s *GRPCServer) GetSchedulesIDs(ctx context.Context, req *pb.UserIDRequest) (*pb.ScheduleIDList, error) {
	s.logger.Info("got getting schedule ids request in grpc",
		slog.Int64("user_id", req.UserId))

	scheduleIDs, err := s.service.GetSchedulesIDs(req.UserId)
	if err != nil {
		s.logger.Error("error getting schedules ids in grpc", slog.String("error", err.Error()))
		return nil, err
	}
	return &pb.ScheduleIDList{
		ScheduleIds: scheduleIDs,
	}, nil
}

func (s *GRPCServer) GetNextTakings(ctx context.Context, req *pb.UserIDRequest) (*pb.TakingList, error) {
	s.logger.Info("got GetNextTakings request in groc",
		slog.Int64("user_id", req.UserId))

	takings, err := s.service.GetNextTakings(req.UserId)

	if err != nil {
		s.logger.Error("error getting next takings in grpc", slog.String("error", err.Error()))
		return nil, err
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

	schedule, err := s.service.GetSchedule(req.UserId, req.ScheduleId)

	if err != nil {
		s.logger.Error("error getting schedule in grpc", slog.String("error", err.Error()))
		return nil, err
	}

	return &pb.ScheduleResponse{
		Id:           schedule.ID,
		MedicineName: schedule.MedicineName,
		StartDate:    schedule.StartDate,
		EndDate:      schedule.EndDate,
		UserId:       schedule.UserID,
		TakingTime:   schedule.TakingTime,
	}, nil
}

func (s *GRPCServer) Run(addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer()
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
