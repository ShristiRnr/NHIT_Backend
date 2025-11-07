package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	authpb "github.com/ShristiRnr/NHIT_Backend/api/proto"
	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a new gRPC-Gateway mux
	mux := runtime.NewServeMux()

	// gRPC service endpoints
	userServiceEndpoint := "localhost:50051"
	authServiceEndpoint := "localhost:50052"
	organizationServiceEndpoint := "localhost:8080"
	departmentServiceEndpoint := "localhost:50054"
	designationServiceEndpoint := "localhost:50055"
	vendorServiceEndpoint := "localhost:50056"

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register User Service
	err := userpb.RegisterUserManagementHandlerFromEndpoint(ctx, mux, userServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register user service gateway: %v", err)
	}
	log.Printf("✅ Registered User Service gateway -> %s", userServiceEndpoint)

	// Register Auth Service
	err = authpb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, authServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register auth service gateway: %v", err)
	}
	log.Printf("✅ Registered Auth Service gateway -> %s", authServiceEndpoint)

	// Register Organization Service
	err = organizationpb.RegisterOrganizationServiceHandlerFromEndpoint(ctx, mux, organizationServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register organization service gateway: %v", err)
	}
	log.Printf("✅ Registered Organization Service gateway -> %s", organizationServiceEndpoint)

	// Register Department Service
	err = departmentpb.RegisterDepartmentServiceHandlerFromEndpoint(ctx, mux, departmentServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register department service gateway: %v", err)
	}
	log.Printf("✅ Registered Department Service gateway -> %s", departmentServiceEndpoint)

	// Register Designation Service
	err = designationpb.RegisterDesignationServiceHandlerFromEndpoint(ctx, mux, designationServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register designation service gateway: %v", err)
	}
	log.Printf("✅ Registered Designation Service gateway -> %s", designationServiceEndpoint)

	// Register Vendor Service
	err = vendorpb.RegisterVendorServiceHandlerFromEndpoint(ctx, mux, vendorServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register vendor service gateway: %v", err)
	}
	log.Printf("✅ Registered Vendor Service gateway -> %s", vendorServiceEndpoint)

	// Add CORS middleware
	handler := cors(mux)

	// Start HTTP server
	port := ":8081"
	log.Printf("API Gateway listening on %s", port)
	log.Printf("REST API available at http://localhost%s/api/v1/", port)
	log.Printf("Examples:")
	log.Printf("   - Users: curl http://localhost%s/api/v1/users", port)
	log.Printf("   - Organizations: curl http://localhost%s/api/v1/organizations", port)
	log.Printf("   - Departments: curl http://localhost%s/api/v1/departments", port)
	log.Printf("   - Designations: curl http://localhost%s/api/v1/designations", port)
	log.Printf("   - Vendors: curl http://localhost%s/api/v1/vendors", port)
	
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// cors adds CORS headers to allow browser requests
func cors(h http.Handler) http.Handler {
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