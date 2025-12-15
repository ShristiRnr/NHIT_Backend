module github.com/ShristiRnr/NHIT_Backend/services/api-gateway

go 1.24.2

require (
	github.com/ShristiRnr/NHIT_Backend/api/pb/authpb v0.0.0-00010101000000-000000000000
	github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb v0.0.0-00010101000000-000000000000
	github.com/ShristiRnr/NHIT_Backend/api/pb/userpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	google.golang.org/grpc v1.77.0
	nhit-note v0.0.0
	nhit-note/api/pb/paymentnotepb v0.0.0-00010101000000-000000000000
)

replace (
	github.com/ShristiRnr/NHIT_Backend/api/pb/authpb => ../../api/pb/authpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb => ../../api/pb/departmentpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb => ../../api/pb/designationpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb => ../../api/pb/organizationpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb => ../../api/pb/projectpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/userpb => ../../api/pb/userpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb => ../../api/pb/vendorpb
	nhit-note => ../../../Nhit-Note
	nhit-note/api/pb/common => ../../../Nhit-Note/api/pb/common
	nhit-note/api/pb/greennotepb => ../../../Nhit-Note/api/pb/greennotepb
	nhit-note/api/pb/paymentnotepb => ../../../Nhit-Note/api/pb/paymentnotepb
	nhit-note/api/pb/paymentpb => ../../../Nhit-Note/api/pb/paymentpb
)

require (
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251124214823-79d6a2a48846 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251111163417-95abcf5c77ba // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
