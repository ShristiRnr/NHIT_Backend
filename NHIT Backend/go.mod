module github.com/ShristiRnr/NHIT_Backend

go 1.24.2

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	google.golang.org/genproto/googleapis/api v0.0.0-20250929231259-57b25ae835d4
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

replace github.com/ShristiRnr/NHIT_Backend/api/pb/userpb => ./api/pb/userpb

replace github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb => ./api/pb/vendorpb

require (
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250929231259-57b25ae835d4 // indirect
)
