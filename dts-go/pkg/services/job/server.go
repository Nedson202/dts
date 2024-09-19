package job

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nedson202/dts-go/internal/job"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/middleware"
	pb "github.com/nedson202/dts-go/proto/job/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

// allowCORS allows Cross Origin Resource Sharing from any origin.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
}

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
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("0.0.0.0:%s", s.grpcPort),
		grpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %v", err)
	}

	gwmux := runtime.NewServeMux()
	err = pb.RegisterJobServiceHandlerClient(context.Background(), gwmux, pb.NewJobServiceClient(conn))
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	corsHandler := allowCORS(gwmux)
	loggedHandler := middleware.LoggingMiddleware(corsHandler)

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.httpPort),
		Handler: loggedHandler,
	}

	logger.Info().Msgf("Starting HTTP server on port %s...", s.httpPort)
	return gwServer.ListenAndServe()
}

