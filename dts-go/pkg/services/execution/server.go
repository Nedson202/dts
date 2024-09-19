package execution

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nedson202/dts-go/internal/execution"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/middleware"
	pb "github.com/nedson202/dts-go/proto/execution/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedExecutionServiceServer
	grpcPort string
	httpPort string
	service  *execution.Service
}

func NewServer(service *execution.Service, grpcPort, httpPort string) *Server {
	return &Server{
		service:  service,
		grpcPort: grpcPort,
		httpPort: httpPort,
	}
}

// Implement the gRPC service methods
func (s *Server) GetExecution(ctx context.Context, req *pb.GetExecutionRequest) (*pb.ExecutionResponse, error) {
	return s.service.GetExecution(ctx, req)
}

func (s *Server) ListExecutions(ctx context.Context, req *pb.ListExecutionsRequest) (*pb.ListExecutionsResponse, error) {
	return s.service.ListExecutions(ctx, req)
}

func (s *Server) Run() error {
	// Create a listener for gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create a gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryServerInterceptor()),
	)
	pb.RegisterExecutionServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	// Start gRPC server
	go func() {
		logger.Info().Msgf("Starting Execution Service gRPC server on port %s", s.grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error().Err(err).Msg("Failed to serve gRPC")
		}
	}()

	// Create a client connection to the gRPC server
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("0.0.0.0:%s", s.grpcPort),
		grpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %v", err)
	}

	gwmux := runtime.NewServeMux()
	err = pb.RegisterExecutionServiceHandlerClient(context.Background(), gwmux, pb.NewExecutionServiceClient(conn))
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	corsHandler := middleware.AllowCORS(gwmux)
	loggedHandler := middleware.LoggingMiddleware(corsHandler)

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.httpPort),
		Handler: loggedHandler,
	}

	logger.Info().Msgf("Starting Execution Service HTTP server on port %s", s.httpPort)
	return gwServer.ListenAndServe()
}
