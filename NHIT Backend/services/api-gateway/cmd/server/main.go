package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	greennotepb "nhit-note/api/pb/greennotepb"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	projectpb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// metadataAnnotator extracts JWT claims and adds them to gRPC metadata
func metadataAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	
	// Extract Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return md
	}

	// Remove "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return md // No "Bearer " prefix found
	}

	// Parse JWT without verification (just to extract claims)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Printf("Failed to parse JWT: %v", err)
		return md
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Add user_id to metadata
		if userID, exists := claims["user_id"]; exists {
			if userIDStr, ok := userID.(string); ok {
				md.Set("user_id", userIDStr)
			}
		}
		
		// Add tenant_id to metadata
		if tenantID, exists := claims["tenant_id"]; exists {
			if tenantIDStr, ok := tenantID.(string); ok {
				md.Set("tenant_id", tenantIDStr)
			}
		}
		
		// Add org_id to metadata
		if orgID, exists := claims["org_id"]; exists {
			if orgIDStr, ok := orgID.(string); ok {
				md.Set("org_id", orgIDStr)
			}
		}
	}

	return md
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a new gRPC-Gateway mux with metadata annotator
	mux := runtime.NewServeMux(
		runtime.WithMetadata(metadataAnnotator),
	)

	// gRPC service endpoints
	userServiceEndpoint := "localhost:50051"
	authServiceEndpoint := "localhost:50052"
	organizationServiceEndpoint := "localhost:8082"
	departmentServiceEndpoint := "localhost:50054"
	designationServiceEndpoint := "localhost:50055"
	vendorServiceEndpoint := "localhost:50058"
	projectServiceEndpoint := "localhost:50057"
	greennoteServiceEndpoint := "localhost:50059"

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
		log.Fatalf("Failed to register organization service: %v", err)
	}
	log.Printf("✅ Registered Organization Service  -> %s", organizationServiceEndpoint)

	// Register Department Service
	err = departmentpb.RegisterDepartmentServiceHandlerFromEndpoint(ctx, mux, departmentServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register department service: %v", err)
	}
	log.Printf("✅ Registered Department Service  -> %s", departmentServiceEndpoint)

	// Register Designation Service
	err = designationpb.RegisterDesignationServiceHandlerFromEndpoint(ctx, mux, designationServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register designation service : %v", err)
	}
	log.Printf("✅ Registered Designation Service  -> %s", designationServiceEndpoint)

	// Register Vendor Service
	err = vendorpb.RegisterVendorServiceHandlerFromEndpoint(ctx, mux, vendorServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register vendor service: %v", err)
	}
	log.Printf("✅ Registered Vendor Service-> %s", vendorServiceEndpoint)

	// Register Project Service
	err = projectpb.RegisterProjectServiceHandlerFromEndpoint(ctx, mux, projectServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register project service: %v", err)
	}
	log.Printf("✅ Registered Project Service -> %s", projectServiceEndpoint)

	err = greennotepb.RegisterGreenNoteServiceHandlerFromEndpoint(ctx, mux, greennoteServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register GreenNote service: %v", err)
	}
	log.Printf("✅ Registered GreenNote Service -> %s", greennoteServiceEndpoint)

	// Add CORS middleware
	handler := cors(mux)

	// Start HTTP server
	port := ":8083"
	log.Printf("API Gateway listening on %s", port)
	log.Printf("REST API available at http://localhost%s/api/v1/", port)
	log.Printf("Examples:")
	log.Printf("   - Users: curl http://localhost%s/api/v1/users", port)
	log.Printf("   - Organizations: curl http://localhost%s/api/v1/organizations", port)
	log.Printf("   - Departments: curl http://localhost%s/api/v1/departments", port)
	log.Printf("   - Designations: curl http://localhost%s/api/v1/designations", port)
	log.Printf("   - Vendors: curl http://localhost%s/api/v1/vendors", port)
	log.Printf("   - Projects: curl http://localhost%s/api/v1/projects", port)
	log.Printf("   - Tenants: curl http://localhost%s/api/v1/tenants", port)
	log.Printf("   - Green Notes: curl http://localhost%s/api/v1/green-notes", port)

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// cors adds CORS headers to allow browser requests and handles authentication forwarding
func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		requestHeaders := r.Header.Get("Access-Control-Request-Headers")
		if requestHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-Tenant-ID, X-Org-ID, tenant_id, org_id")
		}
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Log incoming requests for debugging
		log.Printf("📝 %s %s - Auth header present: %v", r.Method, r.URL.Path, r.Header.Get("Authorization") != "")

		// Forward request to handler
		h.ServeHTTP(w, r)
	})
}
