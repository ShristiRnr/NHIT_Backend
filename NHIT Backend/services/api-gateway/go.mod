module github.com/ShristiRnr/NHIT_Backend/services/api-gateway

go 1.24.2

require (
	github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/userpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/proto v0.0.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	google.golang.org/grpc v1.76.0
)

replace github.com/ShristiRnr/NHIT_Backend/api/proto => ../../api/proto

replace github.com/ShristiRnr/NHIT_Backend/api/pb/userpb => ../../api/pb/userpb

replace github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb => ../../api/pb/departmentpb

replace github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb => ../../api/pb/designationpb

require (
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250929231259-57b25ae835d4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250929231259-57b25ae835d4 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
