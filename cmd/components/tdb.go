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

package components

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	tdb "github.com/milvus-io/milvus/internal/distributed/tdb"
	"github.com/milvus-io/milvus/internal/util/dependency"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/util/paramtable"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// TDB implements TDB grpc server
type TDB struct {
	ctx context.Context
	svr *tdb.Server
}

// NewTDB creates a new TDB component
func NewTDB(ctx context.Context, factory dependency.Factory) (*TDB, error) {
	svr, err := tdb.NewServer(ctx, factory)
	if err != nil {
		return nil, err
	}
	return &TDB{
		ctx: ctx,
		svr: svr,
	}, nil
}

// Prepare prepares the TDB component
func (t *TDB) Prepare() error {
	return t.svr.Prepare()
}

// Run starts the TDB service
func (t *TDB) Run() error {
	if err := t.svr.Run(); err != nil {
		log.Ctx(t.ctx).Error("TDB starts error", zap.Error(err))
		return err
	}
	log.Ctx(t.ctx).Info("TDB successfully started")
	return nil
}

// Stop terminates the TDB service
func (t *TDB) Stop() error {
	timeout := paramtable.Get().TDBCfg.GracefulStopTimeout.GetAsDuration(time.Second)
	return exitWhenStopTimeout(t.svr.Stop, timeout)
}

// Health returns TDB's health status
func (t *TDB) Health(ctx context.Context) commonpb.StateCode {
	// TODO: Implement health check for TDB
	// For now, return healthy if server is running
	return commonpb.StateCode_Healthy
}

// GetName returns the component name
func (t *TDB) GetName() string {
	return typeutil.TDBRole
}

// GetComponentStates returns TDB's component states
func (t *TDB) GetComponentStates(ctx context.Context) *milvuspb.ComponentStates {
	return &milvuspb.ComponentStates{
		State: &milvuspb.ComponentInfo{
			NodeID:    paramtable.GetNodeID(),
			Role:      typeutil.TDBRole,
			StateCode: commonpb.StateCode_Healthy,
		},
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
	}
}
