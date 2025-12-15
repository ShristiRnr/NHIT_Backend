module nhit-note/services/payment-service

go 1.21

require (
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	github.com/lib/pq v1.10.9
	nhit-note/api/pb/paymentpb v0.0.0
)

replace nhit-note/api/pb/paymentpb => ../../api/pb/paymentpb
