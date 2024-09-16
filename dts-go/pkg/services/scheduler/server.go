package scheduler

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nedson202/dts-go/internal/scheduler"
	pb "github.com/nedson202/dts-go/pkg/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedSchedulerServiceServer
	grpcPort string
	httpPort string
	service  *scheduler.Service
}

func NewServer(service *scheduler.Service, grpcPort, httpPort string) *Server {
	return &Server{
		service:  service,
		grpcPort: grpcPort,
		httpPort: httpPort,
	}
}

// Implement the gRPC service methods
func (s *Server) ScheduleJob(ctx context.Context, req *pb.ScheduleJobRequest) (*pb.ScheduleJobResponse, error) {
	return s.service.ScheduleJob(ctx, req)
}

func (s *Server) CancelJob(ctx context.Context, req *pb.CancelJobRequest) (*pb.CancelJobResponse, error) {
	return s.service.CancelJob(ctx, req)
}

func (s *Server) GetScheduledJob(ctx context.Context, req *pb.GetScheduledJobRequest) (*pb.ScheduledJob, error) {
	return s.service.GetScheduledJob(ctx, req)
}

func (s *Server) ListScheduledJobs(ctx context.Context, req *pb.ListScheduledJobsRequest) (*pb.ListScheduledJobsResponse, error) {
	return s.service.ListScheduledJobs(ctx, req)
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func (s *Server) Run() error {
	// Create a listener for gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create a gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterSchedulerServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	// Start gRPC server
	go func() {
		fmt.Printf("Starting gRPC server on port %s...\n", s.grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Printf("Failed to serve gRPC: %v\n", err)
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
	err = pb.RegisterSchedulerServiceHandlerClient(context.Background(), gwmux, pb.NewSchedulerServiceClient(conn))
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	corsHandler := allowCORS(gwmux)
	loggedHandler := loggingMiddleware(corsHandler)

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.httpPort),
		Handler: loggedHandler,
	}

	fmt.Printf("Starting HTTP server on port %s...\n", s.httpPort)
	return gwServer.ListenAndServe()
}

func (s *Server) RegisterWithGRPCServer(grpcServer *grpc.Server) {
	pb.RegisterSchedulerServiceServer(grpcServer, s)
}
