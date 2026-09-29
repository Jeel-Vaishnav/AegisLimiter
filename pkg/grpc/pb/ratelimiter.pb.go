package ratelimiter_pb

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type AlgorithmType int32

const (
	AlgorithmType_ALGORITHM_UNSPECIFIED   AlgorithmType = 0
	AlgorithmType_TOKEN_BUCKET             AlgorithmType = 1
	AlgorithmType_SLIDING_WINDOW_COUNTER   AlgorithmType = 2
	AlgorithmType_SLIDING_WINDOW_LOG       AlgorithmType = 3
	AlgorithmType_LEAKY_BUCKET             AlgorithmType = 4
)

func (x AlgorithmType) String() string {
	switch x {
	case AlgorithmType_TOKEN_BUCKET:
		return "TOKEN_BUCKET"
	case AlgorithmType_SLIDING_WINDOW_COUNTER:
		return "SLIDING_WINDOW_COUNTER"
	case AlgorithmType_SLIDING_WINDOW_LOG:
		return "SLIDING_WINDOW_LOG"
	case AlgorithmType_LEAKY_BUCKET:
		return "LEAKY_BUCKET"
	default:
		return "ALGORITHM_UNSPECIFIED"
	}
}

type RateLimitRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key           string        `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Algorithm     AlgorithmType `protobuf:"varint,2,opt,name=algorithm,proto3,enum=ratelimiter.v1.AlgorithmType" json:"algorithm,omitempty"`
	Cost          int64         `protobuf:"varint,3,opt,name=cost,proto3" json:"cost,omitempty"`
	Capacity      int64         `protobuf:"varint,4,opt,name=capacity,proto3" json:"capacity,omitempty"`
	RatePerSecond float64       `protobuf:"fixed64,5,opt,name=rate_per_second,json=ratePerSecond,proto3" json:"rate_per_second,omitempty"`
	WindowSizeMs  int64         `protobuf:"varint,6,opt,name=window_size_ms,json=windowSizeMs,proto3" json:"window_size_ms,omitempty"`
}

func (x *RateLimitRequest) Reset()         { *x = RateLimitRequest{} }
func (x *RateLimitRequest) String() string { return x.Key }
func (*RateLimitRequest) ProtoMessage()    {}
func (x *RateLimitRequest) ProtoReflect() protoreflect.Message {
	return nil
}

func (x *RateLimitRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *RateLimitRequest) GetAlgorithm() AlgorithmType {
	if x != nil {
		return x.Algorithm
	}
	return AlgorithmType_ALGORITHM_UNSPECIFIED
}

func (x *RateLimitRequest) GetCost() int64 {
	if x != nil {
		return x.Cost
	}
	return 1
}

func (x *RateLimitRequest) GetCapacity() int64 {
	if x != nil {
		return x.Capacity
	}
	return 0
}

func (x *RateLimitRequest) GetRatePerSecond() float64 {
	if x != nil {
		return x.RatePerSecond
	}
	return 0
}

func (x *RateLimitRequest) GetWindowSizeMs() int64 {
	if x != nil {
		return x.WindowSizeMs
	}
	return 0
}

type RateLimitResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Allowed      bool          `protobuf:"varint,1,opt,name=allowed,proto3" json:"allowed,omitempty"`
	Remaining    int64         `protobuf:"varint,2,opt,name=remaining,proto3" json:"remaining,omitempty"`
	Limit        int64         `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	RetryAfterMs int64         `protobuf:"varint,4,opt,name=retry_after_ms,json=retryAfterMs,proto3" json:"retry_after_ms,omitempty"`
	ResetMs      int64         `protobuf:"varint,5,opt,name=reset_ms,json=resetMs,proto3" json:"reset_ms,omitempty"`
	Algorithm    AlgorithmType `protobuf:"varint,6,opt,name=algorithm,proto3,enum=ratelimiter.v1.AlgorithmType" json:"algorithm,omitempty"`
	CacheTier    string        `protobuf:"bytes,7,opt,name=cache_tier,json=cacheTier,proto3" json:"cache_tier,omitempty"`
}

func (x *RateLimitResponse) Reset()         { *x = RateLimitResponse{} }
func (x *RateLimitResponse) String() string { return "" }
func (*RateLimitResponse) ProtoMessage()    {}
func (x *RateLimitResponse) ProtoReflect() protoreflect.Message {
	return nil
}

type StatusRequest struct {
	Key       string        `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Algorithm AlgorithmType `protobuf:"varint,2,opt,name=algorithm,proto3,enum=ratelimiter.v1.AlgorithmType" json:"algorithm,omitempty"`
}

func (x *StatusRequest) Reset()         { *x = StatusRequest{} }
func (x *StatusRequest) String() string { return "" }
func (*StatusRequest) ProtoMessage()    {}

type StatusResponse struct {
	Key       string        `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Algorithm AlgorithmType `protobuf:"varint,2,opt,name=algorithm,proto3,enum=ratelimiter.v1.AlgorithmType" json:"algorithm,omitempty"`
	Remaining int64         `protobuf:"varint,3,opt,name=remaining,proto3" json:"remaining,omitempty"`
	Limit     int64         `protobuf:"varint,4,opt,name=limit,proto3" json:"limit,omitempty"`
	ResetMs   int64         `protobuf:"varint,5,opt,name=reset_ms,json=resetMs,proto3" json:"reset_ms,omitempty"`
}

func (x *StatusResponse) Reset()         { *x = StatusResponse{} }
func (x *StatusResponse) String() string { return "" }
func (*StatusResponse) ProtoMessage()    {}

type ResetRequest struct {
	Key       string        `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Algorithm AlgorithmType `protobuf:"varint,2,opt,name=algorithm,proto3,enum=ratelimiter.v1.AlgorithmType" json:"algorithm,omitempty"`
}

func (x *ResetRequest) Reset()         { *x = ResetRequest{} }
func (x *ResetRequest) String() string { return "" }
func (*ResetRequest) ProtoMessage()    {}

type ResetResponse struct {
	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ResetResponse) Reset()         { *x = ResetResponse{} }
func (x *ResetResponse) String() string { return "" }
func (*ResetResponse) ProtoMessage()    {}

type ConfigureRuleRequest struct {
	ClientTier    string        `protobuf:"bytes,1,opt,name=client_tier,json=clientTier,proto3" json:"client_tier,omitempty"`
	Algorithm     AlgorithmType `protobuf:"varint,2,opt,name=algorithm,proto3,enum=ratelimiter.v1.AlgorithmType" json:"algorithm,omitempty"`
	Capacity      int64         `protobuf:"varint,3,opt,name=capacity,proto3" json:"capacity,omitempty"`
	RatePerSecond float64       `protobuf:"fixed64,4,opt,name=rate_per_second,json=ratePerSecond,proto3" json:"rate_per_second,omitempty"`
	WindowSizeMs  int64         `protobuf:"varint,5,opt,name=window_size_ms,json=windowSizeMs,proto3" json:"window_size_ms,omitempty"`
}

func (x *ConfigureRuleRequest) Reset()         { *x = ConfigureRuleRequest{} }
func (x *ConfigureRuleRequest) String() string { return "" }
func (*ConfigureRuleRequest) ProtoMessage()    {}

type ConfigureRuleResponse struct {
	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ConfigureRuleResponse) Reset()         { *x = ConfigureRuleResponse{} }
func (x *ConfigureRuleResponse) String() string { return "" }
func (*ConfigureRuleResponse) ProtoMessage()    {}

// RateLimiterServiceClient is the client API for RateLimiterService service.
type RateLimiterServiceClient interface {
	CheckRateLimit(ctx context.Context, in *RateLimitRequest, opts ...grpc.CallOption) (*RateLimitResponse, error)
	GetStatus(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	ResetLimit(ctx context.Context, in *ResetRequest, opts ...grpc.CallOption) (*ResetResponse, error)
	ConfigureRule(ctx context.Context, in *ConfigureRuleRequest, opts ...grpc.CallOption) (*ConfigureRuleResponse, error)
}

type rateLimiterServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRateLimiterServiceClient(cc grpc.ClientConnInterface) RateLimiterServiceClient {
	return &rateLimiterServiceClient{cc}
}

func (c *rateLimiterServiceClient) CheckRateLimit(ctx context.Context, in *RateLimitRequest, opts ...grpc.CallOption) (*RateLimitResponse, error) {
	out := new(RateLimitResponse)
	err := c.cc.Invoke(ctx, "/ratelimiter.v1.RateLimiterService/CheckRateLimit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rateLimiterServiceClient) GetStatus(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, "/ratelimiter.v1.RateLimiterService/GetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rateLimiterServiceClient) ResetLimit(ctx context.Context, in *ResetRequest, opts ...grpc.CallOption) (*ResetResponse, error) {
	out := new(ResetResponse)
	err := c.cc.Invoke(ctx, "/ratelimiter.v1.RateLimiterService/ResetLimit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rateLimiterServiceClient) ConfigureRule(ctx context.Context, in *ConfigureRuleRequest, opts ...grpc.CallOption) (*ConfigureRuleResponse, error) {
	out := new(ConfigureRuleResponse)
	err := c.cc.Invoke(ctx, "/ratelimiter.v1.RateLimiterService/ConfigureRule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RateLimiterServiceServer is the server API for RateLimiterService service.
type RateLimiterServiceServer interface {
	CheckRateLimit(context.Context, *RateLimitRequest) (*RateLimitResponse, error)
	GetStatus(context.Context, *StatusRequest) (*StatusResponse, error)
	ResetLimit(context.Context, *ResetRequest) (*ResetResponse, error)
	ConfigureRule(context.Context, *ConfigureRuleRequest) (*ConfigureRuleResponse, error)
	mustEmbedUnimplementedRateLimiterServiceServer()
}

type UnimplementedRateLimiterServiceServer struct{}

func (UnimplementedRateLimiterServiceServer) CheckRateLimit(context.Context, *RateLimitRequest) (*RateLimitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckRateLimit not implemented")
}
func (UnimplementedRateLimiterServiceServer) GetStatus(context.Context, *StatusRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (UnimplementedRateLimiterServiceServer) ResetLimit(context.Context, *ResetRequest) (*ResetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetLimit not implemented")
}
func (UnimplementedRateLimiterServiceServer) ConfigureRule(context.Context, *ConfigureRuleRequest) (*ConfigureRuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfigureRule not implemented")
}
func (UnimplementedRateLimiterServiceServer) mustEmbedUnimplementedRateLimiterServiceServer() {}

func RegisterRateLimiterServiceServer(s grpc.ServiceRegistrar, srv RateLimiterServiceServer) {
	s.RegisterService(&RateLimiterService_ServiceDesc, srv)
}

var RateLimiterService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ratelimiter.v1.RateLimiterService",
	HandlerType: (*RateLimiterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckRateLimit",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(RateLimitRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(RateLimiterServiceServer).CheckRateLimit(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/ratelimiter.v1.RateLimiterService/CheckRateLimit",
				}
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return srv.(RateLimiterServiceServer).CheckRateLimit(ctx, req.(*RateLimitRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
		{
			MethodName: "GetStatus",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(StatusRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(RateLimiterServiceServer).GetStatus(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/ratelimiter.v1.RateLimiterService/GetStatus",
				}
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return srv.(RateLimiterServiceServer).GetStatus(ctx, req.(*StatusRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
		{
			MethodName: "ResetLimit",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(ResetRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(RateLimiterServiceServer).ResetLimit(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/ratelimiter.v1.RateLimiterService/ResetLimit",
				}
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return srv.(RateLimiterServiceServer).ResetLimit(ctx, req.(*ResetRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
		{
			MethodName: "ConfigureRule",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(ConfigureRuleRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(RateLimiterServiceServer).ConfigureRule(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/ratelimiter.v1.RateLimiterService/ConfigureRule",
				}
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return srv.(RateLimiterServiceServer).ConfigureRule(ctx, req.(*ConfigureRuleRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ratelimiter.proto",
}
