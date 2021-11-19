// Autogenerated by Thrift Compiler (facebook)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
// @generated

package storage

import (
	"bytes"
	"context"
	"sync"
	"fmt"
	thrift "github.com/facebook/fbthrift/thrift/lib/go/thrift"
	nebula0 "github.com/vesoft-inc/nebula-go/v2/nebula"
	meta1 "github.com/vesoft-inc/nebula-go/v2/nebula/meta"

)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = sync.Mutex{}
var _ = bytes.Equal
var _ = context.Background

var _ = nebula0.GoUnusedProtection__
var _ = meta1.GoUnusedProtection__
type InternalStorageService interface {
  // Parameters:
  //  - Req
  ChainAddEdges(ctx context.Context, req *ChainAddEdgesRequest) (_r *ExecResponse, err error)
  // Parameters:
  //  - Req
  ChainUpdateEdge(ctx context.Context, req *ChainUpdateEdgeRequest) (_r *UpdateResponse, err error)
}

type InternalStorageServiceClientInterface interface {
  thrift.ClientInterface
  // Parameters:
  //  - Req
  ChainAddEdges(req *ChainAddEdgesRequest) (_r *ExecResponse, err error)
  // Parameters:
  //  - Req
  ChainUpdateEdge(req *ChainUpdateEdgeRequest) (_r *UpdateResponse, err error)
}

type InternalStorageServiceClient struct {
  InternalStorageServiceClientInterface
  CC thrift.ClientConn
}

func(client *InternalStorageServiceClient) Open() error {
  return client.CC.Open()
}

func(client *InternalStorageServiceClient) Close() error {
  return client.CC.Close()
}

func(client *InternalStorageServiceClient) IsOpen() bool {
  return client.CC.IsOpen()
}

func NewInternalStorageServiceClientFactory(t thrift.Transport, f thrift.ProtocolFactory) *InternalStorageServiceClient {
  return &InternalStorageServiceClient{ CC: thrift.NewClientConn(t, f) }
}

func NewInternalStorageServiceClient(t thrift.Transport, iprot thrift.Protocol, oprot thrift.Protocol) *InternalStorageServiceClient {
  return &InternalStorageServiceClient{ CC: thrift.NewClientConnWithProtocols(t, iprot, oprot) }
}

func NewInternalStorageServiceClientProtocol(prot thrift.Protocol) *InternalStorageServiceClient {
  return NewInternalStorageServiceClient(prot.Transport(), prot, prot)
}

// Parameters:
//  - Req
func (p *InternalStorageServiceClient) ChainAddEdges(req *ChainAddEdgesRequest) (_r *ExecResponse, err error) {
  args := InternalStorageServiceChainAddEdgesArgs{
    Req : req,
  }
  err = p.CC.SendMsg("chainAddEdges", &args, thrift.CALL)
  if err != nil { return }
  return p.recvChainAddEdges()
}


func (p *InternalStorageServiceClient) recvChainAddEdges() (value *ExecResponse, err error) {
  var result InternalStorageServiceChainAddEdgesResult
  err = p.CC.RecvMsg("chainAddEdges", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *InternalStorageServiceClient) ChainUpdateEdge(req *ChainUpdateEdgeRequest) (_r *UpdateResponse, err error) {
  args := InternalStorageServiceChainUpdateEdgeArgs{
    Req : req,
  }
  err = p.CC.SendMsg("chainUpdateEdge", &args, thrift.CALL)
  if err != nil { return }
  return p.recvChainUpdateEdge()
}


func (p *InternalStorageServiceClient) recvChainUpdateEdge() (value *UpdateResponse, err error) {
  var result InternalStorageServiceChainUpdateEdgeResult
  err = p.CC.RecvMsg("chainUpdateEdge", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}


type InternalStorageServiceThreadsafeClient struct {
  InternalStorageServiceClientInterface
  CC thrift.ClientConn
  Mu sync.Mutex
}

func(client *InternalStorageServiceThreadsafeClient) Open() error {
  client.Mu.Lock()
  defer client.Mu.Unlock()
  return client.CC.Open()
}

func(client *InternalStorageServiceThreadsafeClient) Close() error {
  client.Mu.Lock()
  defer client.Mu.Unlock()
  return client.CC.Close()
}

func(client *InternalStorageServiceThreadsafeClient) IsOpen() bool {
  client.Mu.Lock()
  defer client.Mu.Unlock()
  return client.CC.IsOpen()
}

func NewInternalStorageServiceThreadsafeClientFactory(t thrift.Transport, f thrift.ProtocolFactory) *InternalStorageServiceThreadsafeClient {
  return &InternalStorageServiceThreadsafeClient{ CC: thrift.NewClientConn(t, f) }
}

func NewInternalStorageServiceThreadsafeClient(t thrift.Transport, iprot thrift.Protocol, oprot thrift.Protocol) *InternalStorageServiceThreadsafeClient {
  return &InternalStorageServiceThreadsafeClient{ CC: thrift.NewClientConnWithProtocols(t, iprot, oprot) }
}

func NewInternalStorageServiceThreadsafeClientProtocol(prot thrift.Protocol) *InternalStorageServiceThreadsafeClient {
  return NewInternalStorageServiceThreadsafeClient(prot.Transport(), prot, prot)
}

// Parameters:
//  - Req
func (p *InternalStorageServiceThreadsafeClient) ChainAddEdges(req *ChainAddEdgesRequest) (_r *ExecResponse, err error) {
  p.Mu.Lock()
  defer p.Mu.Unlock()
  args := InternalStorageServiceChainAddEdgesArgs{
    Req : req,
  }
  err = p.CC.SendMsg("chainAddEdges", &args, thrift.CALL)
  if err != nil { return }
  return p.recvChainAddEdges()
}


func (p *InternalStorageServiceThreadsafeClient) recvChainAddEdges() (value *ExecResponse, err error) {
  var result InternalStorageServiceChainAddEdgesResult
  err = p.CC.RecvMsg("chainAddEdges", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *InternalStorageServiceThreadsafeClient) ChainUpdateEdge(req *ChainUpdateEdgeRequest) (_r *UpdateResponse, err error) {
  p.Mu.Lock()
  defer p.Mu.Unlock()
  args := InternalStorageServiceChainUpdateEdgeArgs{
    Req : req,
  }
  err = p.CC.SendMsg("chainUpdateEdge", &args, thrift.CALL)
  if err != nil { return }
  return p.recvChainUpdateEdge()
}


func (p *InternalStorageServiceThreadsafeClient) recvChainUpdateEdge() (value *UpdateResponse, err error) {
  var result InternalStorageServiceChainUpdateEdgeResult
  err = p.CC.RecvMsg("chainUpdateEdge", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}


type InternalStorageServiceChannelClient struct {
  RequestChannel thrift.RequestChannel
}

func (c *InternalStorageServiceChannelClient) Close() error {
  return c.RequestChannel.Close()
}

func (c *InternalStorageServiceChannelClient) IsOpen() bool {
  return c.RequestChannel.IsOpen()
}

func (c *InternalStorageServiceChannelClient) Open() error {
  return c.RequestChannel.Open()
}

func NewInternalStorageServiceChannelClient(channel thrift.RequestChannel) *InternalStorageServiceChannelClient {
  return &InternalStorageServiceChannelClient{RequestChannel: channel}
}

// Parameters:
//  - Req
func (p *InternalStorageServiceChannelClient) ChainAddEdges(ctx context.Context, req *ChainAddEdgesRequest) (_r *ExecResponse, err error) {
  args := InternalStorageServiceChainAddEdgesArgs{
    Req : req,
  }
  var result InternalStorageServiceChainAddEdgesResult
  err = p.RequestChannel.Call(ctx, "chainAddEdges", &args, &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *InternalStorageServiceChannelClient) ChainUpdateEdge(ctx context.Context, req *ChainUpdateEdgeRequest) (_r *UpdateResponse, err error) {
  args := InternalStorageServiceChainUpdateEdgeArgs{
    Req : req,
  }
  var result InternalStorageServiceChainUpdateEdgeResult
  err = p.RequestChannel.Call(ctx, "chainUpdateEdge", &args, &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}


type InternalStorageServiceProcessor struct {
  processorMap map[string]thrift.ProcessorFunctionContext
  handler InternalStorageService
}

func (p *InternalStorageServiceProcessor) AddToProcessorMap(key string, processor thrift.ProcessorFunctionContext) {
  p.processorMap[key] = processor
}

func (p *InternalStorageServiceProcessor) GetProcessorFunctionContext(key string) (processor thrift.ProcessorFunctionContext, err error) {
  if processor, ok := p.processorMap[key]; ok {
    return processor, nil
  }
  return nil, nil // generic error message will be sent
}

func (p *InternalStorageServiceProcessor) ProcessorMap() map[string]thrift.ProcessorFunctionContext {
  return p.processorMap
}

func NewInternalStorageServiceProcessor(handler InternalStorageService) *InternalStorageServiceProcessor {
  self311 := &InternalStorageServiceProcessor{handler:handler, processorMap:make(map[string]thrift.ProcessorFunctionContext)}
  self311.processorMap["chainAddEdges"] = &internalStorageServiceProcessorChainAddEdges{handler:handler}
  self311.processorMap["chainUpdateEdge"] = &internalStorageServiceProcessorChainUpdateEdge{handler:handler}
  return self311
}

type internalStorageServiceProcessorChainAddEdges struct {
  handler InternalStorageService
}

func (p *internalStorageServiceProcessorChainAddEdges) Read(iprot thrift.Protocol) (thrift.Struct, thrift.Exception) {
  args := InternalStorageServiceChainAddEdgesArgs{}
  if err := args.Read(iprot); err != nil {
    return nil, err
  }
  iprot.ReadMessageEnd()
  return &args, nil
}

func (p *internalStorageServiceProcessorChainAddEdges) Write(seqId int32, result thrift.WritableStruct, oprot thrift.Protocol) (err thrift.Exception) {
  var err2 error
  messageType := thrift.REPLY
  switch result.(type) {
  case thrift.ApplicationException:
    messageType = thrift.EXCEPTION
  }
  if err2 = oprot.WriteMessageBegin("chainAddEdges", messageType, seqId); err2 != nil {
    err = err2
  }
  if err2 = result.Write(oprot); err == nil && err2 != nil {
    err = err2
  }
  if err2 = oprot.WriteMessageEnd(); err == nil && err2 != nil {
    err = err2
  }
  if err2 = oprot.Flush(); err == nil && err2 != nil {
    err = err2
  }
  return err
}

func (p *internalStorageServiceProcessorChainAddEdges) RunContext(ctx context.Context, argStruct thrift.Struct) (thrift.WritableStruct, thrift.ApplicationException) {
  args := argStruct.(*InternalStorageServiceChainAddEdgesArgs)
  var result InternalStorageServiceChainAddEdgesResult
  if retval, err := p.handler.ChainAddEdges(ctx, args.Req); err != nil {
    switch err.(type) {
    default:
      x := thrift.NewApplicationException(thrift.INTERNAL_ERROR, "Internal error processing chainAddEdges: " + err.Error())
      return x, x
    }
  } else {
    result.Success = retval
  }
  return &result, nil
}

type internalStorageServiceProcessorChainUpdateEdge struct {
  handler InternalStorageService
}

func (p *internalStorageServiceProcessorChainUpdateEdge) Read(iprot thrift.Protocol) (thrift.Struct, thrift.Exception) {
  args := InternalStorageServiceChainUpdateEdgeArgs{}
  if err := args.Read(iprot); err != nil {
    return nil, err
  }
  iprot.ReadMessageEnd()
  return &args, nil
}

func (p *internalStorageServiceProcessorChainUpdateEdge) Write(seqId int32, result thrift.WritableStruct, oprot thrift.Protocol) (err thrift.Exception) {
  var err2 error
  messageType := thrift.REPLY
  switch result.(type) {
  case thrift.ApplicationException:
    messageType = thrift.EXCEPTION
  }
  if err2 = oprot.WriteMessageBegin("chainUpdateEdge", messageType, seqId); err2 != nil {
    err = err2
  }
  if err2 = result.Write(oprot); err == nil && err2 != nil {
    err = err2
  }
  if err2 = oprot.WriteMessageEnd(); err == nil && err2 != nil {
    err = err2
  }
  if err2 = oprot.Flush(); err == nil && err2 != nil {
    err = err2
  }
  return err
}

func (p *internalStorageServiceProcessorChainUpdateEdge) RunContext(ctx context.Context, argStruct thrift.Struct) (thrift.WritableStruct, thrift.ApplicationException) {
  args := argStruct.(*InternalStorageServiceChainUpdateEdgeArgs)
  var result InternalStorageServiceChainUpdateEdgeResult
  if retval, err := p.handler.ChainUpdateEdge(ctx, args.Req); err != nil {
    switch err.(type) {
    default:
      x := thrift.NewApplicationException(thrift.INTERNAL_ERROR, "Internal error processing chainUpdateEdge: " + err.Error())
      return x, x
    }
  } else {
    result.Success = retval
  }
  return &result, nil
}


// HELPER FUNCTIONS AND STRUCTURES

// Attributes:
//  - Req
type InternalStorageServiceChainAddEdgesArgs struct {
  thrift.IRequest
  Req *ChainAddEdgesRequest `thrift:"req,1" db:"req" json:"req"`
}

func NewInternalStorageServiceChainAddEdgesArgs() *InternalStorageServiceChainAddEdgesArgs {
  return &InternalStorageServiceChainAddEdgesArgs{
    Req: NewChainAddEdgesRequest(),
  }
}

var InternalStorageServiceChainAddEdgesArgs_Req_DEFAULT *ChainAddEdgesRequest
func (p *InternalStorageServiceChainAddEdgesArgs) GetReq() *ChainAddEdgesRequest {
  if !p.IsSetReq() {
    return InternalStorageServiceChainAddEdgesArgs_Req_DEFAULT
  }
return p.Req
}
func (p *InternalStorageServiceChainAddEdgesArgs) IsSetReq() bool {
  return p != nil && p.Req != nil
}

func (p *InternalStorageServiceChainAddEdgesArgs) Read(iprot thrift.Protocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if err := p.ReadField1(iprot); err != nil {
        return err
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *InternalStorageServiceChainAddEdgesArgs)  ReadField1(iprot thrift.Protocol) error {
  p.Req = NewChainAddEdgesRequest()
  if err := p.Req.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *InternalStorageServiceChainAddEdgesArgs) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("chainAddEdges_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField1(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *InternalStorageServiceChainAddEdgesArgs) writeField1(oprot thrift.Protocol) (err error) {
  if err := oprot.WriteFieldBegin("req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *InternalStorageServiceChainAddEdgesArgs) String() string {
  if p == nil {
    return "<nil>"
  }

  var reqVal string
  if p.Req == nil {
    reqVal = "<nil>"
  } else {
    reqVal = fmt.Sprintf("%v", p.Req)
  }
  return fmt.Sprintf("InternalStorageServiceChainAddEdgesArgs({Req:%s})", reqVal)
}

// Attributes:
//  - Success
type InternalStorageServiceChainAddEdgesResult struct {
  thrift.IResponse
  Success *ExecResponse `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewInternalStorageServiceChainAddEdgesResult() *InternalStorageServiceChainAddEdgesResult {
  return &InternalStorageServiceChainAddEdgesResult{}
}

var InternalStorageServiceChainAddEdgesResult_Success_DEFAULT *ExecResponse
func (p *InternalStorageServiceChainAddEdgesResult) GetSuccess() *ExecResponse {
  if !p.IsSetSuccess() {
    return InternalStorageServiceChainAddEdgesResult_Success_DEFAULT
  }
return p.Success
}
func (p *InternalStorageServiceChainAddEdgesResult) IsSetSuccess() bool {
  return p != nil && p.Success != nil
}

func (p *InternalStorageServiceChainAddEdgesResult) Read(iprot thrift.Protocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 0:
      if err := p.ReadField0(iprot); err != nil {
        return err
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *InternalStorageServiceChainAddEdgesResult)  ReadField0(iprot thrift.Protocol) error {
  p.Success = NewExecResponse()
  if err := p.Success.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *InternalStorageServiceChainAddEdgesResult) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("chainAddEdges_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField0(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *InternalStorageServiceChainAddEdgesResult) writeField0(oprot thrift.Protocol) (err error) {
  if p.IsSetSuccess() {
    if err := oprot.WriteFieldBegin("success", thrift.STRUCT, 0); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err) }
    if err := p.Success.Write(oprot); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
    }
    if err := oprot.WriteFieldEnd(); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err) }
  }
  return err
}

func (p *InternalStorageServiceChainAddEdgesResult) String() string {
  if p == nil {
    return "<nil>"
  }

  var successVal string
  if p.Success == nil {
    successVal = "<nil>"
  } else {
    successVal = fmt.Sprintf("%v", p.Success)
  }
  return fmt.Sprintf("InternalStorageServiceChainAddEdgesResult({Success:%s})", successVal)
}

// Attributes:
//  - Req
type InternalStorageServiceChainUpdateEdgeArgs struct {
  thrift.IRequest
  Req *ChainUpdateEdgeRequest `thrift:"req,1" db:"req" json:"req"`
}

func NewInternalStorageServiceChainUpdateEdgeArgs() *InternalStorageServiceChainUpdateEdgeArgs {
  return &InternalStorageServiceChainUpdateEdgeArgs{
    Req: NewChainUpdateEdgeRequest(),
  }
}

var InternalStorageServiceChainUpdateEdgeArgs_Req_DEFAULT *ChainUpdateEdgeRequest
func (p *InternalStorageServiceChainUpdateEdgeArgs) GetReq() *ChainUpdateEdgeRequest {
  if !p.IsSetReq() {
    return InternalStorageServiceChainUpdateEdgeArgs_Req_DEFAULT
  }
return p.Req
}
func (p *InternalStorageServiceChainUpdateEdgeArgs) IsSetReq() bool {
  return p != nil && p.Req != nil
}

func (p *InternalStorageServiceChainUpdateEdgeArgs) Read(iprot thrift.Protocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if err := p.ReadField1(iprot); err != nil {
        return err
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *InternalStorageServiceChainUpdateEdgeArgs)  ReadField1(iprot thrift.Protocol) error {
  p.Req = NewChainUpdateEdgeRequest()
  if err := p.Req.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *InternalStorageServiceChainUpdateEdgeArgs) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("chainUpdateEdge_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField1(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *InternalStorageServiceChainUpdateEdgeArgs) writeField1(oprot thrift.Protocol) (err error) {
  if err := oprot.WriteFieldBegin("req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *InternalStorageServiceChainUpdateEdgeArgs) String() string {
  if p == nil {
    return "<nil>"
  }

  var reqVal string
  if p.Req == nil {
    reqVal = "<nil>"
  } else {
    reqVal = fmt.Sprintf("%v", p.Req)
  }
  return fmt.Sprintf("InternalStorageServiceChainUpdateEdgeArgs({Req:%s})", reqVal)
}

// Attributes:
//  - Success
type InternalStorageServiceChainUpdateEdgeResult struct {
  thrift.IResponse
  Success *UpdateResponse `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewInternalStorageServiceChainUpdateEdgeResult() *InternalStorageServiceChainUpdateEdgeResult {
  return &InternalStorageServiceChainUpdateEdgeResult{}
}

var InternalStorageServiceChainUpdateEdgeResult_Success_DEFAULT *UpdateResponse
func (p *InternalStorageServiceChainUpdateEdgeResult) GetSuccess() *UpdateResponse {
  if !p.IsSetSuccess() {
    return InternalStorageServiceChainUpdateEdgeResult_Success_DEFAULT
  }
return p.Success
}
func (p *InternalStorageServiceChainUpdateEdgeResult) IsSetSuccess() bool {
  return p != nil && p.Success != nil
}

func (p *InternalStorageServiceChainUpdateEdgeResult) Read(iprot thrift.Protocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 0:
      if err := p.ReadField0(iprot); err != nil {
        return err
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *InternalStorageServiceChainUpdateEdgeResult)  ReadField0(iprot thrift.Protocol) error {
  p.Success = NewUpdateResponse()
  if err := p.Success.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *InternalStorageServiceChainUpdateEdgeResult) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("chainUpdateEdge_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField0(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *InternalStorageServiceChainUpdateEdgeResult) writeField0(oprot thrift.Protocol) (err error) {
  if p.IsSetSuccess() {
    if err := oprot.WriteFieldBegin("success", thrift.STRUCT, 0); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err) }
    if err := p.Success.Write(oprot); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
    }
    if err := oprot.WriteFieldEnd(); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err) }
  }
  return err
}

func (p *InternalStorageServiceChainUpdateEdgeResult) String() string {
  if p == nil {
    return "<nil>"
  }

  var successVal string
  if p.Success == nil {
    successVal = "<nil>"
  } else {
    successVal = fmt.Sprintf("%v", p.Success)
  }
  return fmt.Sprintf("InternalStorageServiceChainUpdateEdgeResult({Success:%s})", successVal)
}


