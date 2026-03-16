// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tdb

import (
	"context"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/milvus-io/milvus/internal/agentcoord"
	"github.com/milvus-io/milvus/internal/eventcoord"
	"github.com/milvus-io/milvus/internal/memorycoord"
	"github.com/milvus-io/milvus/internal/util/dependency"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/proto/agentpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/eventpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/memorypb"
	"github.com/milvus-io/milvus/pkg/v2/util/interceptor"
	"github.com/milvus-io/milvus/pkg/v2/util/logutil"
	"github.com/milvus-io/milvus/pkg/v2/util/netutil"
	"github.com/milvus-io/milvus/pkg/v2/util/paramtable"
)

// Server is the TDB gRPC server that wraps all TDB services.
type Server struct {
	agentServer  *AgentServer
	memoryServer *MemoryServer
	eventServer  *EventServer

	grpcServer *grpc.Server
	listener   *netutil.NetListener

	agentCoord  *agentcoord.AgentCoord
	memoryCoord *memorycoord.MemoryCoord
	eventCoord  *eventcoord.EventCoord

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	serverID atomic.Int64
}

// NewServer creates a new TDB server.
func NewServer(ctx context.Context, factory dependency.Factory) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)

	s := &Server{
		ctx:    ctx,
		cancel: cancel,
	}

	var err error

	// Initialize coordinators
	s.agentCoord, err = agentcoord.NewAgentCoord(ctx, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	s.memoryCoord, err = memorycoord.NewMemoryCoord(ctx, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	s.eventCoord, err = eventcoord.NewEventCoord(ctx, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	// Initialize gRPC servers
	s.agentServer = NewAgentServer(s.agentCoord)
	s.memoryServer = NewMemoryServer(s.memoryCoord)
	s.eventServer = NewEventServer(s.eventCoord)

	return s, nil
}

// Prepare prepares the server.
func (s *Server) Prepare() error {
	log := log.Ctx(s.ctx)

	listener, err := netutil.NewListener(
		netutil.OptIP(paramtable.Get().TDBGrpcServerCfg.IP),
		netutil.OptPort(paramtable.Get().TDBGrpcServerCfg.Port.GetAsInt()),
	)
	if err != nil {
		log.Warn("TDB fail to create net listener", zap.Error(err))
		return err
	}
	log.Info("TDB listen on", zap.String("address", listener.Addr().String()), zap.Int("port", listener.Port()))
	s.listener = listener
	return nil
}

// Run initializes and starts the TDB server.
func (s *Server) Run() error {
	if err := s.init(); err != nil {
		return err
	}
	log.Ctx(s.ctx).Info("TDB init done ...")

	if err := s.start(); err != nil {
		return err
	}
	log.Ctx(s.ctx).Info("TDB start done ...")
	return nil
}

func (s *Server) init() error {
	log := log.Ctx(s.ctx)

	// Initialize coordinators
	if err := s.agentCoord.Init(); err != nil {
		log.Error("failed to init agent coord", zap.Error(err))
		return err
	}

	if err := s.memoryCoord.Init(); err != nil {
		log.Error("failed to init memory coord", zap.Error(err))
		return err
	}

	if err := s.eventCoord.Init(); err != nil {
		log.Error("failed to init event coord", zap.Error(err))
		return err
	}

	// Start gRPC server
	if err := s.startGrpc(); err != nil {
		return err
	}

	return nil
}

func (s *Server) startGrpc() error {
	s.wg.Add(1)
	go s.startGrpcLoop()
	return nil
}

func (s *Server) startGrpcLoop() {
	defer s.wg.Done()

	Params := &paramtable.Get().TDBGrpcServerCfg
	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	kasp := keepalive.ServerParameters{
		Time:    60 * time.Second,
		Timeout: 10 * time.Second,
	}

	log := log.Ctx(s.ctx)
	log.Info("start TDB grpc ", zap.Int("port", s.listener.Port()))

	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	grpcOpts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.MaxRecvMsgSize(Params.ServerMaxRecvSize.GetAsInt()),
		grpc.MaxSendMsgSize(Params.ServerMaxSendSize.GetAsInt()),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			logutil.UnaryTraceLoggerInterceptor,
			interceptor.ClusterValidationUnaryServerInterceptor(),
			interceptor.ServerIDValidationUnaryServerInterceptor(func() int64 {
				if s.serverID.Load() == 0 {
					s.serverID.Store(paramtable.GetNodeID())
				}
				return s.serverID.Load()
			}),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			logutil.StreamTraceLoggerInterceptor,
			interceptor.ClusterValidationStreamServerInterceptor(),
			interceptor.ServerIDValidationStreamServerInterceptor(func() int64 {
				if s.serverID.Load() == 0 {
					s.serverID.Store(paramtable.GetNodeID())
				}
				return s.serverID.Load()
			}),
		)),
	}

	s.grpcServer = grpc.NewServer(grpcOpts...)

	// Register TDB services
	agentpb.RegisterAgentServiceServer(s.grpcServer, s.agentServer)
	memorypb.RegisterMemoryServiceServer(s.grpcServer, s.memoryServer)
	eventpb.RegisterEventServiceServer(s.grpcServer, s.eventServer)

	log.Info("TDB gRPC services registered")

	if err := s.grpcServer.Serve(s.listener); err != nil {
		log.Error("TDB gRPC server error", zap.Error(err))
	}
}

func (s *Server) start() error {
	log := log.Ctx(s.ctx)
	log.Info("TDB coordinators starting ...")

	// Start coordinators
	if err := s.agentCoord.Start(); err != nil {
		log.Error("failed to start agent coord", zap.Error(err))
		return err
	}

	if err := s.memoryCoord.Start(); err != nil {
		log.Error("failed to start memory coord", zap.Error(err))
		return err
	}

	if err := s.eventCoord.Start(); err != nil {
		log.Error("failed to start event coord", zap.Error(err))
		return err
	}

	return nil
}

// Stop stops the TDB server.
func (s *Server) Stop() error {
	logger := log.Ctx(s.ctx)
	if s.listener != nil {
		logger = logger.With(zap.String("address", s.listener.Address()))
	}
	logger.Info("TDB stopping")
	defer func() {
		logger.Info("TDB stopped")
	}()

	// Stop coordinators
	if s.agentCoord != nil {
		if err := s.agentCoord.Stop(); err != nil {
			logger.Error("failed to stop agent coord", zap.Error(err))
		}
	}

	if s.memoryCoord != nil {
		if err := s.memoryCoord.Stop(); err != nil {
			logger.Error("failed to stop memory coord", zap.Error(err))
		}
	}

	if s.eventCoord != nil {
		if err := s.eventCoord.Stop(); err != nil {
			logger.Error("failed to stop event coord", zap.Error(err))
		}
	}

	// Stop gRPC server
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	s.wg.Wait()
	s.cancel()

	if s.listener != nil {
		s.listener.Close()
	}

	return nil
}

// GetAgentCoord returns the agent coordinator.
func (s *Server) GetAgentCoord() *agentcoord.AgentCoord {
	return s.agentCoord
}

// GetMemoryCoord returns the memory coordinator.
func (s *Server) GetMemoryCoord() *memorycoord.MemoryCoord {
	return s.memoryCoord
}

// GetEventCoord returns the event coordinator.
func (s *Server) GetEventCoord() *eventcoord.EventCoord {
	return s.eventCoord
}
