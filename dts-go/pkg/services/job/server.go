package job

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nedson202/dts-go/internal/job"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/middleware"
	pb "github.com/nedson202/dts-go/proto/job/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	pb.UnimplementedJobServiceServer
	grpcPort string
	httpPort string
	service  *job.Service
}

func NewServer(service *job.Service, grpcPort, httpPort string) *Server {
	return &Server{
		service:  service,
		grpcPort: grpcPort,
		httpPort: httpPort,
	}
}

// Implement the gRPC service methods
func (s *Server) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	return s.service.CreateJob(ctx, req)
}

func (s *Server) GetJob(ctx context.Context, req *pb.GetJobRequest) (*pb.JobResponse, error) {
	return s.service.GetJob(ctx, req)
}

func (s *Server) ListJobs(ctx context.Context, req *pb.ListJobsRequest) (*pb.ListJobsResponse, error) {
	return s.service.ListJobs(ctx, req)
}

func (s *Server) UpdateJob(ctx context.Context, req *pb.UpdateJobRequest) (*pb.JobResponse, error) {
	return s.service.UpdateJob(ctx, req)
}

func (s *Server) DeleteJob(ctx context.Context, req *pb.DeleteJobRequest) (*pb.DeleteJobResponse, error) {
	return s.service.DeleteJob(ctx, req)
}

func (s *Server) CancelJob(ctx context.Context, req *pb.CancelJobRequest) (*pb.CancelJobResponse, error) {
	return s.service.CancelJob(ctx, req)
}

// Implement the HTTP service methods
func (s *Server) Run() error {
	// Create a listener for gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create a gRPC server with logging interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryServerInterceptor()),
	)
	pb.RegisterJobServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	// Start gRPC server
	go func() {
		logger.Info().Msgf("Starting gRPC server on port %s...", s.grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error().Err(err).Msg("Failed to serve gRPC")
		}
	}()

	// Create a client connection to the gRPC server
	ctx := context.Background()
	conn, err := grpc.NewClient(ctx, fmt.Sprintf("0.0.0.0:%s", s.grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC client: %v", err)
	}
	defer conn.Close()

	gwmux := runtime.NewServeMux()
	err = pb.RegisterJobServiceHandlerClient(ctx, gwmux, pb.NewJobServiceClient(conn))
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	corsHandler := middleware.AllowCORS(gwmux)
	loggedHandler := middleware.LoggingMiddleware(corsHandler)

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.httpPort),
		Handler: loggedHandler,
	}

	logger.Info().Msgf("Starting HTTP server on port %s...", s.httpPort)
	return gwServer.ListenAndServe()
}

