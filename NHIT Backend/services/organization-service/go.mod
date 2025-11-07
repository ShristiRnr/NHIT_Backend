module github.com/ShristiRnr/NHIT_Backend/services/organization-service

go 1.24.2

require (
	github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/services/shared v0.0.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

replace github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb => ../../api/pb/organizationpb

replace github.com/ShristiRnr/NHIT_Backend/services/shared => ../shared

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/lib/pq v1.10.9 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250929231259-57b25ae835d4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250929231259-57b25ae835d4 // indirect
)
