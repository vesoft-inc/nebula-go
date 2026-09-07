// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package nebula_ng

import (
	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/graph"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

type (
	summary struct {
		summary *graph.Summary
	}
	planInfo struct {
		planInfo *graph.PlanInfo
	}
	queryStats struct {
		queryStats *graph.QueryStats
	}
)

func (s *summary) ParseTimeUs() int64 {
	return s.summary.ElapsedTime.ParseTimeUs
}

func (s *summary) BuildTimeUs() int64 {
	return s.summary.ElapsedTime.BuildTimeUs
}

func (s *summary) OptimizeTimeUs() int64 {
	return s.summary.ElapsedTime.OptimizeTimeUs
}

func (s *summary) ExecutionTimeUs() int64 {
	return s.summary.ElapsedTime.ExecutionTimeUs
}

func (s *summary) SerializeTimeUs() int64 {
	return s.summary.ElapsedTime.SerializeTimeUs
}

func (s *summary) TotalServerTimeUs() int64 {
	return s.summary.ElapsedTime.TotalServerTimeUs
}

func (s *summary) ExplainType() string {
	return string(s.summary.ExplainType)
}

func (s *summary) PlanInfo() types.PlanInfo {
	return &planInfo{s.summary.GetPlanInfo()}
}

func (s *summary) QueryStats() types.QueryStats {
	return &queryStats{s.summary.GetQueryStats()}
}

func (s *summary) Log() string {
	return string(s.summary.LogStream)
}

func (s *summary) NumWarnings() int {
	return int(s.summary.NumWarnings)
}

func (p *planInfo) Id() string {
	return string(p.planInfo.Id)
}

func (p *planInfo) Name() string {
	return string(p.planInfo.Name)
}

func (p *planInfo) Details() string {
	return string(p.planInfo.Details)
}

func (p *planInfo) Columns() []string {
	columnsVec := make([]string, 0, len(p.planInfo.Columns))
	for _, column := range p.planInfo.Columns {
		columnsVec = append(columnsVec, string(column))
	}
	return columnsVec
}

func (p *planInfo) TimeMs() float64 {
	return p.planInfo.TimeMs
}

func (p *planInfo) Rows() int64 {
	return p.planInfo.Rows
}

func (p *planInfo) MemoryKib() float64 {
	return p.planInfo.MemoryKib
}

func (p *planInfo) BlockedMs() float64 {
	return p.planInfo.BlockedMs
}

func (p *planInfo) QueuedMs() float64 {
	return p.planInfo.QueuedMs
}

func (p *planInfo) ConsumeMs() float64 {
	return p.planInfo.ConsumeMs
}

func (p *planInfo) ProduceMs() float64 {
	return p.planInfo.ProduceMs
}

func (p *planInfo) FinishMs() float64 {
	return p.planInfo.FinishMs
}

func (p *planInfo) Batches() int64 {
	return p.planInfo.Batches
}

func (p *planInfo) Concurrency() int64 {
	return p.planInfo.Concurrency
}

func (p *planInfo) OtherStatsJson() []byte {
	return p.planInfo.OtherStatsJson
}

func (p *planInfo) Children() []types.PlanInfo {
	children := make([]types.PlanInfo, 0, len(p.planInfo.Children))
	for _, child := range p.planInfo.Children {
		children = append(children, &planInfo{planInfo: child})
	}
	return children
}

func (q *queryStats) NumAffectedNodes() int64 {
	return q.queryStats.NumAffectedNodes
}

func (q *queryStats) NumAffectedEdges() int64 {
	return q.queryStats.NumAffectedEdges
}

func (q *queryStats) ExportedRowSize() int {
	return int(q.queryStats.NumExportedRecords)
}

func (q *queryStats) ExportedPaths() []string {
	paths := make([]string, 0, len(q.queryStats.ExportedPaths))
	for _, path := range q.queryStats.ExportedPaths {
		paths = append(paths, string(path))
	}
	return paths
}
