package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/handlers"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = ":50056"
	httpPort = ":8086"
)

func main() {
	ctx := context.Background()

	// Initialize production-level handler
	vendorHandler := handlers.NewVendorHandler()

	// Start gRPC server
	go func() {
		if err := startGRPCServer(vendorHandler); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start HTTP gateway server
	go func() {
		if err := startHTTPGateway(ctx); err != nil {
			log.Fatalf("Failed to start HTTP gateway: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Shutting down servers...")
}

func startGRPCServer(vendorHandler *handlers.VendorHandler) error {
	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	server := grpcLib.NewServer()
	
	// Register vendor handler - handles all vendor and account operations
	vendorpb.RegisterVendorServiceServer(server, vendorHandler)

	// Enable reflection for development
	reflection.Register(server)

	log.Printf("🚀 gRPC server listening on %s", grpcPort)
	log.Printf("📋 Production-level vendor service with complete business logic")
	return server.Serve(listener)
}

func startHTTPGateway(ctx context.Context) error {
	mux := runtime.NewServeMux()

	opts := []grpcLib.DialOption{grpcLib.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := fmt.Sprintf("localhost%s", grpcPort)

	err := vendorpb.RegisterVendorServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Add CORS middleware
	handler := corsMiddleware(mux)

	log.Printf("🌐 HTTP gateway listening on %s", httpPort)
	log.Printf("📋 Production-level API endpoints available:")
	log.Printf("   - POST   /api/v1/vendors                    # Create vendor")
	log.Printf("   - GET    /api/v1/vendors/{id}               # Get vendor by ID")
	log.Printf("   - GET    /api/v1/vendors/code/{code}        # Get vendor by code")
	log.Printf("   - PUT    /api/v1/vendors/{id}               # Update vendor")
	log.Printf("   - DELETE /api/v1/vendors/{id}               # Delete vendor")
	log.Printf("   - GET    /api/v1/vendors                    # List vendors")
	log.Printf("   - POST   /api/v1/vendors/generate-code      # Generate vendor code")
	log.Printf("   - PUT    /api/v1/vendors/{id}/code          # Update vendor code")
	log.Printf("   - POST   /api/v1/vendors/{id}/regenerate-code # Regenerate vendor code")
	log.Printf("   - POST   /api/v1/vendors/{id}/accounts      # Create vendor account")
	log.Printf("   - GET    /api/v1/vendors/{id}/accounts      # Get vendor accounts")
	log.Printf("   - GET    /api/v1/vendors/{id}/banking-details # Get banking details")
	log.Printf("   - PUT    /api/v1/vendors/accounts/{id}      # Update vendor account")
	log.Printf("   - DELETE /api/v1/vendors/accounts/{id}      # Delete vendor account")
	log.Printf("   - POST   /api/v1/vendors/accounts/{id}/toggle-status # Toggle account status")
	
	return http.ListenAndServe(httpPort, handler)
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
