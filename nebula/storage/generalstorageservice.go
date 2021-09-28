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
type GeneralStorageService interface {
  // Parameters:
  //  - Req
  Get(ctx context.Context, req *KVGetRequest) (_r *KVGetResponse, err error)
  // Parameters:
  //  - Req
  Put(ctx context.Context, req *KVPutRequest) (_r *ExecResponse, err error)
  // Parameters:
  //  - Req
  Remove(ctx context.Context, req *KVRemoveRequest) (_r *ExecResponse, err error)
}

type GeneralStorageServiceClientInterface interface {
  thrift.ClientInterface
  // Parameters:
  //  - Req
  Get(req *KVGetRequest) (_r *KVGetResponse, err error)
  // Parameters:
  //  - Req
  Put(req *KVPutRequest) (_r *ExecResponse, err error)
  // Parameters:
  //  - Req
  Remove(req *KVRemoveRequest) (_r *ExecResponse, err error)
}

type GeneralStorageServiceClient struct {
  GeneralStorageServiceClientInterface
  CC thrift.ClientConn
}

func(client *GeneralStorageServiceClient) Open() error {
  return client.CC.Open()
}

func(client *GeneralStorageServiceClient) Close() error {
  return client.CC.Close()
}

func(client *GeneralStorageServiceClient) IsOpen() bool {
  return client.CC.IsOpen()
}

func NewGeneralStorageServiceClientFactory(t thrift.Transport, f thrift.ProtocolFactory) *GeneralStorageServiceClient {
  return &GeneralStorageServiceClient{ CC: thrift.NewClientConn(t, f) }
}

func NewGeneralStorageServiceClient(t thrift.Transport, iprot thrift.Protocol, oprot thrift.Protocol) *GeneralStorageServiceClient {
  return &GeneralStorageServiceClient{ CC: thrift.NewClientConnWithProtocols(t, iprot, oprot) }
}

func NewGeneralStorageServiceClientProtocol(prot thrift.Protocol) *GeneralStorageServiceClient {
  return NewGeneralStorageServiceClient(prot.Transport(), prot, prot)
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceClient) Get(req *KVGetRequest) (_r *KVGetResponse, err error) {
  args := GeneralStorageServiceGetArgs{
    Req : req,
  }
  err = p.CC.SendMsg("get", &args, thrift.CALL)
  if err != nil { return }
  return p.recvGet()
}


func (p *GeneralStorageServiceClient) recvGet() (value *KVGetResponse, err error) {
  var result GeneralStorageServiceGetResult
  err = p.CC.RecvMsg("get", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceClient) Put(req *KVPutRequest) (_r *ExecResponse, err error) {
  args := GeneralStorageServicePutArgs{
    Req : req,
  }
  err = p.CC.SendMsg("put", &args, thrift.CALL)
  if err != nil { return }
  return p.recvPut()
}


func (p *GeneralStorageServiceClient) recvPut() (value *ExecResponse, err error) {
  var result GeneralStorageServicePutResult
  err = p.CC.RecvMsg("put", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceClient) Remove(req *KVRemoveRequest) (_r *ExecResponse, err error) {
  args := GeneralStorageServiceRemoveArgs{
    Req : req,
  }
  err = p.CC.SendMsg("remove", &args, thrift.CALL)
  if err != nil { return }
  return p.recvRemove()
}


func (p *GeneralStorageServiceClient) recvRemove() (value *ExecResponse, err error) {
  var result GeneralStorageServiceRemoveResult
  err = p.CC.RecvMsg("remove", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}


type GeneralStorageServiceThreadsafeClient struct {
  GeneralStorageServiceClientInterface
  CC thrift.ClientConn
  Mu sync.Mutex
}

func(client *GeneralStorageServiceThreadsafeClient) Open() error {
  client.Mu.Lock()
  defer client.Mu.Unlock()
  return client.CC.Open()
}

func(client *GeneralStorageServiceThreadsafeClient) Close() error {
  client.Mu.Lock()
  defer client.Mu.Unlock()
  return client.CC.Close()
}

func(client *GeneralStorageServiceThreadsafeClient) IsOpen() bool {
  client.Mu.Lock()
  defer client.Mu.Unlock()
  return client.CC.IsOpen()
}

func NewGeneralStorageServiceThreadsafeClientFactory(t thrift.Transport, f thrift.ProtocolFactory) *GeneralStorageServiceThreadsafeClient {
  return &GeneralStorageServiceThreadsafeClient{ CC: thrift.NewClientConn(t, f) }
}

func NewGeneralStorageServiceThreadsafeClient(t thrift.Transport, iprot thrift.Protocol, oprot thrift.Protocol) *GeneralStorageServiceThreadsafeClient {
  return &GeneralStorageServiceThreadsafeClient{ CC: thrift.NewClientConnWithProtocols(t, iprot, oprot) }
}

func NewGeneralStorageServiceThreadsafeClientProtocol(prot thrift.Protocol) *GeneralStorageServiceThreadsafeClient {
  return NewGeneralStorageServiceThreadsafeClient(prot.Transport(), prot, prot)
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceThreadsafeClient) Get(req *KVGetRequest) (_r *KVGetResponse, err error) {
  p.Mu.Lock()
  defer p.Mu.Unlock()
  args := GeneralStorageServiceGetArgs{
    Req : req,
  }
  err = p.CC.SendMsg("get", &args, thrift.CALL)
  if err != nil { return }
  return p.recvGet()
}


func (p *GeneralStorageServiceThreadsafeClient) recvGet() (value *KVGetResponse, err error) {
  var result GeneralStorageServiceGetResult
  err = p.CC.RecvMsg("get", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceThreadsafeClient) Put(req *KVPutRequest) (_r *ExecResponse, err error) {
  p.Mu.Lock()
  defer p.Mu.Unlock()
  args := GeneralStorageServicePutArgs{
    Req : req,
  }
  err = p.CC.SendMsg("put", &args, thrift.CALL)
  if err != nil { return }
  return p.recvPut()
}


func (p *GeneralStorageServiceThreadsafeClient) recvPut() (value *ExecResponse, err error) {
  var result GeneralStorageServicePutResult
  err = p.CC.RecvMsg("put", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceThreadsafeClient) Remove(req *KVRemoveRequest) (_r *ExecResponse, err error) {
  p.Mu.Lock()
  defer p.Mu.Unlock()
  args := GeneralStorageServiceRemoveArgs{
    Req : req,
  }
  err = p.CC.SendMsg("remove", &args, thrift.CALL)
  if err != nil { return }
  return p.recvRemove()
}


func (p *GeneralStorageServiceThreadsafeClient) recvRemove() (value *ExecResponse, err error) {
  var result GeneralStorageServiceRemoveResult
  err = p.CC.RecvMsg("remove", &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}


type GeneralStorageServiceChannelClient struct {
  RequestChannel thrift.RequestChannel
}

func (c *GeneralStorageServiceChannelClient) Close() error {
  return c.RequestChannel.Close()
}

func (c *GeneralStorageServiceChannelClient) IsOpen() bool {
  return c.RequestChannel.IsOpen()
}

func (c *GeneralStorageServiceChannelClient) Open() error {
  return c.RequestChannel.Open()
}

func NewGeneralStorageServiceChannelClient(channel thrift.RequestChannel) *GeneralStorageServiceChannelClient {
  return &GeneralStorageServiceChannelClient{RequestChannel: channel}
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceChannelClient) Get(ctx context.Context, req *KVGetRequest) (_r *KVGetResponse, err error) {
  args := GeneralStorageServiceGetArgs{
    Req : req,
  }
  var result GeneralStorageServiceGetResult
  err = p.RequestChannel.Call(ctx, "get", &args, &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceChannelClient) Put(ctx context.Context, req *KVPutRequest) (_r *ExecResponse, err error) {
  args := GeneralStorageServicePutArgs{
    Req : req,
  }
  var result GeneralStorageServicePutResult
  err = p.RequestChannel.Call(ctx, "put", &args, &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}

// Parameters:
//  - Req
func (p *GeneralStorageServiceChannelClient) Remove(ctx context.Context, req *KVRemoveRequest) (_r *ExecResponse, err error) {
  args := GeneralStorageServiceRemoveArgs{
    Req : req,
  }
  var result GeneralStorageServiceRemoveResult
  err = p.RequestChannel.Call(ctx, "remove", &args, &result)
  if err != nil { return }

  return result.GetSuccess(), nil
}


type GeneralStorageServiceProcessor struct {
  processorMap map[string]thrift.ProcessorFunctionContext
  handler GeneralStorageService
}

func (p *GeneralStorageServiceProcessor) AddToProcessorMap(key string, processor thrift.ProcessorFunctionContext) {
  p.processorMap[key] = processor
}

func (p *GeneralStorageServiceProcessor) GetProcessorFunctionContext(key string) (processor thrift.ProcessorFunctionContext, err error) {
  if processor, ok := p.processorMap[key]; ok {
    return processor, nil
  }
  return nil, nil // generic error message will be sent
}

func (p *GeneralStorageServiceProcessor) ProcessorMap() map[string]thrift.ProcessorFunctionContext {
  return p.processorMap
}

func NewGeneralStorageServiceProcessor(handler GeneralStorageService) *GeneralStorageServiceProcessor {
  self282 := &GeneralStorageServiceProcessor{handler:handler, processorMap:make(map[string]thrift.ProcessorFunctionContext)}
  self282.processorMap["get"] = &generalStorageServiceProcessorGet{handler:handler}
  self282.processorMap["put"] = &generalStorageServiceProcessorPut{handler:handler}
  self282.processorMap["remove"] = &generalStorageServiceProcessorRemove{handler:handler}
  return self282
}

type generalStorageServiceProcessorGet struct {
  handler GeneralStorageService
}

func (p *generalStorageServiceProcessorGet) Read(iprot thrift.Protocol) (thrift.Struct, thrift.Exception) {
  args := GeneralStorageServiceGetArgs{}
  if err := args.Read(iprot); err != nil {
    return nil, err
  }
  iprot.ReadMessageEnd()
  return &args, nil
}

func (p *generalStorageServiceProcessorGet) Write(seqId int32, result thrift.WritableStruct, oprot thrift.Protocol) (err thrift.Exception) {
  var err2 error
  messageType := thrift.REPLY
  switch result.(type) {
  case thrift.ApplicationException:
    messageType = thrift.EXCEPTION
  }
  if err2 = oprot.WriteMessageBegin("get", messageType, seqId); err2 != nil {
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

func (p *generalStorageServiceProcessorGet) RunContext(ctx context.Context, argStruct thrift.Struct) (thrift.WritableStruct, thrift.ApplicationException) {
  args := argStruct.(*GeneralStorageServiceGetArgs)
  var result GeneralStorageServiceGetResult
  if retval, err := p.handler.Get(ctx, args.Req); err != nil {
    switch err.(type) {
    default:
      x := thrift.NewApplicationException(thrift.INTERNAL_ERROR, "Internal error processing get: " + err.Error())
      return x, x
    }
  } else {
    result.Success = retval
  }
  return &result, nil
}

type generalStorageServiceProcessorPut struct {
  handler GeneralStorageService
}

func (p *generalStorageServiceProcessorPut) Read(iprot thrift.Protocol) (thrift.Struct, thrift.Exception) {
  args := GeneralStorageServicePutArgs{}
  if err := args.Read(iprot); err != nil {
    return nil, err
  }
  iprot.ReadMessageEnd()
  return &args, nil
}

func (p *generalStorageServiceProcessorPut) Write(seqId int32, result thrift.WritableStruct, oprot thrift.Protocol) (err thrift.Exception) {
  var err2 error
  messageType := thrift.REPLY
  switch result.(type) {
  case thrift.ApplicationException:
    messageType = thrift.EXCEPTION
  }
  if err2 = oprot.WriteMessageBegin("put", messageType, seqId); err2 != nil {
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

func (p *generalStorageServiceProcessorPut) RunContext(ctx context.Context, argStruct thrift.Struct) (thrift.WritableStruct, thrift.ApplicationException) {
  args := argStruct.(*GeneralStorageServicePutArgs)
  var result GeneralStorageServicePutResult
  if retval, err := p.handler.Put(ctx, args.Req); err != nil {
    switch err.(type) {
    default:
      x := thrift.NewApplicationException(thrift.INTERNAL_ERROR, "Internal error processing put: " + err.Error())
      return x, x
    }
  } else {
    result.Success = retval
  }
  return &result, nil
}

type generalStorageServiceProcessorRemove struct {
  handler GeneralStorageService
}

func (p *generalStorageServiceProcessorRemove) Read(iprot thrift.Protocol) (thrift.Struct, thrift.Exception) {
  args := GeneralStorageServiceRemoveArgs{}
  if err := args.Read(iprot); err != nil {
    return nil, err
  }
  iprot.ReadMessageEnd()
  return &args, nil
}

func (p *generalStorageServiceProcessorRemove) Write(seqId int32, result thrift.WritableStruct, oprot thrift.Protocol) (err thrift.Exception) {
  var err2 error
  messageType := thrift.REPLY
  switch result.(type) {
  case thrift.ApplicationException:
    messageType = thrift.EXCEPTION
  }
  if err2 = oprot.WriteMessageBegin("remove", messageType, seqId); err2 != nil {
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

func (p *generalStorageServiceProcessorRemove) RunContext(ctx context.Context, argStruct thrift.Struct) (thrift.WritableStruct, thrift.ApplicationException) {
  args := argStruct.(*GeneralStorageServiceRemoveArgs)
  var result GeneralStorageServiceRemoveResult
  if retval, err := p.handler.Remove(ctx, args.Req); err != nil {
    switch err.(type) {
    default:
      x := thrift.NewApplicationException(thrift.INTERNAL_ERROR, "Internal error processing remove: " + err.Error())
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
type GeneralStorageServiceGetArgs struct {
  thrift.IRequest
  Req *KVGetRequest `thrift:"req,1" db:"req" json:"req"`
}

func NewGeneralStorageServiceGetArgs() *GeneralStorageServiceGetArgs {
  return &GeneralStorageServiceGetArgs{
    Req: NewKVGetRequest(),
  }
}

var GeneralStorageServiceGetArgs_Req_DEFAULT *KVGetRequest
func (p *GeneralStorageServiceGetArgs) GetReq() *KVGetRequest {
  if !p.IsSetReq() {
    return GeneralStorageServiceGetArgs_Req_DEFAULT
  }
return p.Req
}
func (p *GeneralStorageServiceGetArgs) IsSetReq() bool {
  return p != nil && p.Req != nil
}

func (p *GeneralStorageServiceGetArgs) Read(iprot thrift.Protocol) error {
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

func (p *GeneralStorageServiceGetArgs)  ReadField1(iprot thrift.Protocol) error {
  p.Req = NewKVGetRequest()
  if err := p.Req.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *GeneralStorageServiceGetArgs) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("get_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField1(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GeneralStorageServiceGetArgs) writeField1(oprot thrift.Protocol) (err error) {
  if err := oprot.WriteFieldBegin("req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *GeneralStorageServiceGetArgs) String() string {
  if p == nil {
    return "<nil>"
  }

  var reqVal string
  if p.Req == nil {
    reqVal = "<nil>"
  } else {
    reqVal = fmt.Sprintf("%v", p.Req)
  }
  return fmt.Sprintf("GeneralStorageServiceGetArgs({Req:%s})", reqVal)
}

// Attributes:
//  - Success
type GeneralStorageServiceGetResult struct {
  thrift.IResponse
  Success *KVGetResponse `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewGeneralStorageServiceGetResult() *GeneralStorageServiceGetResult {
  return &GeneralStorageServiceGetResult{}
}

var GeneralStorageServiceGetResult_Success_DEFAULT *KVGetResponse
func (p *GeneralStorageServiceGetResult) GetSuccess() *KVGetResponse {
  if !p.IsSetSuccess() {
    return GeneralStorageServiceGetResult_Success_DEFAULT
  }
return p.Success
}
func (p *GeneralStorageServiceGetResult) IsSetSuccess() bool {
  return p != nil && p.Success != nil
}

func (p *GeneralStorageServiceGetResult) Read(iprot thrift.Protocol) error {
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

func (p *GeneralStorageServiceGetResult)  ReadField0(iprot thrift.Protocol) error {
  p.Success = NewKVGetResponse()
  if err := p.Success.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *GeneralStorageServiceGetResult) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("get_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField0(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GeneralStorageServiceGetResult) writeField0(oprot thrift.Protocol) (err error) {
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

func (p *GeneralStorageServiceGetResult) String() string {
  if p == nil {
    return "<nil>"
  }

  var successVal string
  if p.Success == nil {
    successVal = "<nil>"
  } else {
    successVal = fmt.Sprintf("%v", p.Success)
  }
  return fmt.Sprintf("GeneralStorageServiceGetResult({Success:%s})", successVal)
}

// Attributes:
//  - Req
type GeneralStorageServicePutArgs struct {
  thrift.IRequest
  Req *KVPutRequest `thrift:"req,1" db:"req" json:"req"`
}

func NewGeneralStorageServicePutArgs() *GeneralStorageServicePutArgs {
  return &GeneralStorageServicePutArgs{
    Req: NewKVPutRequest(),
  }
}

var GeneralStorageServicePutArgs_Req_DEFAULT *KVPutRequest
func (p *GeneralStorageServicePutArgs) GetReq() *KVPutRequest {
  if !p.IsSetReq() {
    return GeneralStorageServicePutArgs_Req_DEFAULT
  }
return p.Req
}
func (p *GeneralStorageServicePutArgs) IsSetReq() bool {
  return p != nil && p.Req != nil
}

func (p *GeneralStorageServicePutArgs) Read(iprot thrift.Protocol) error {
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

func (p *GeneralStorageServicePutArgs)  ReadField1(iprot thrift.Protocol) error {
  p.Req = NewKVPutRequest()
  if err := p.Req.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *GeneralStorageServicePutArgs) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("put_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField1(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GeneralStorageServicePutArgs) writeField1(oprot thrift.Protocol) (err error) {
  if err := oprot.WriteFieldBegin("req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *GeneralStorageServicePutArgs) String() string {
  if p == nil {
    return "<nil>"
  }

  var reqVal string
  if p.Req == nil {
    reqVal = "<nil>"
  } else {
    reqVal = fmt.Sprintf("%v", p.Req)
  }
  return fmt.Sprintf("GeneralStorageServicePutArgs({Req:%s})", reqVal)
}

// Attributes:
//  - Success
type GeneralStorageServicePutResult struct {
  thrift.IResponse
  Success *ExecResponse `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewGeneralStorageServicePutResult() *GeneralStorageServicePutResult {
  return &GeneralStorageServicePutResult{}
}

var GeneralStorageServicePutResult_Success_DEFAULT *ExecResponse
func (p *GeneralStorageServicePutResult) GetSuccess() *ExecResponse {
  if !p.IsSetSuccess() {
    return GeneralStorageServicePutResult_Success_DEFAULT
  }
return p.Success
}
func (p *GeneralStorageServicePutResult) IsSetSuccess() bool {
  return p != nil && p.Success != nil
}

func (p *GeneralStorageServicePutResult) Read(iprot thrift.Protocol) error {
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

func (p *GeneralStorageServicePutResult)  ReadField0(iprot thrift.Protocol) error {
  p.Success = NewExecResponse()
  if err := p.Success.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *GeneralStorageServicePutResult) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("put_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField0(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GeneralStorageServicePutResult) writeField0(oprot thrift.Protocol) (err error) {
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

func (p *GeneralStorageServicePutResult) String() string {
  if p == nil {
    return "<nil>"
  }

  var successVal string
  if p.Success == nil {
    successVal = "<nil>"
  } else {
    successVal = fmt.Sprintf("%v", p.Success)
  }
  return fmt.Sprintf("GeneralStorageServicePutResult({Success:%s})", successVal)
}

// Attributes:
//  - Req
type GeneralStorageServiceRemoveArgs struct {
  thrift.IRequest
  Req *KVRemoveRequest `thrift:"req,1" db:"req" json:"req"`
}

func NewGeneralStorageServiceRemoveArgs() *GeneralStorageServiceRemoveArgs {
  return &GeneralStorageServiceRemoveArgs{
    Req: NewKVRemoveRequest(),
  }
}

var GeneralStorageServiceRemoveArgs_Req_DEFAULT *KVRemoveRequest
func (p *GeneralStorageServiceRemoveArgs) GetReq() *KVRemoveRequest {
  if !p.IsSetReq() {
    return GeneralStorageServiceRemoveArgs_Req_DEFAULT
  }
return p.Req
}
func (p *GeneralStorageServiceRemoveArgs) IsSetReq() bool {
  return p != nil && p.Req != nil
}

func (p *GeneralStorageServiceRemoveArgs) Read(iprot thrift.Protocol) error {
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

func (p *GeneralStorageServiceRemoveArgs)  ReadField1(iprot thrift.Protocol) error {
  p.Req = NewKVRemoveRequest()
  if err := p.Req.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *GeneralStorageServiceRemoveArgs) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("remove_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField1(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GeneralStorageServiceRemoveArgs) writeField1(oprot thrift.Protocol) (err error) {
  if err := oprot.WriteFieldBegin("req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *GeneralStorageServiceRemoveArgs) String() string {
  if p == nil {
    return "<nil>"
  }

  var reqVal string
  if p.Req == nil {
    reqVal = "<nil>"
  } else {
    reqVal = fmt.Sprintf("%v", p.Req)
  }
  return fmt.Sprintf("GeneralStorageServiceRemoveArgs({Req:%s})", reqVal)
}

// Attributes:
//  - Success
type GeneralStorageServiceRemoveResult struct {
  thrift.IResponse
  Success *ExecResponse `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewGeneralStorageServiceRemoveResult() *GeneralStorageServiceRemoveResult {
  return &GeneralStorageServiceRemoveResult{}
}

var GeneralStorageServiceRemoveResult_Success_DEFAULT *ExecResponse
func (p *GeneralStorageServiceRemoveResult) GetSuccess() *ExecResponse {
  if !p.IsSetSuccess() {
    return GeneralStorageServiceRemoveResult_Success_DEFAULT
  }
return p.Success
}
func (p *GeneralStorageServiceRemoveResult) IsSetSuccess() bool {
  return p != nil && p.Success != nil
}

func (p *GeneralStorageServiceRemoveResult) Read(iprot thrift.Protocol) error {
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

func (p *GeneralStorageServiceRemoveResult)  ReadField0(iprot thrift.Protocol) error {
  p.Success = NewExecResponse()
  if err := p.Success.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *GeneralStorageServiceRemoveResult) Write(oprot thrift.Protocol) error {
  if err := oprot.WriteStructBegin("remove_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if err := p.writeField0(oprot); err != nil { return err }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GeneralStorageServiceRemoveResult) writeField0(oprot thrift.Protocol) (err error) {
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

func (p *GeneralStorageServiceRemoveResult) String() string {
  if p == nil {
    return "<nil>"
  }

  var successVal string
  if p.Success == nil {
    successVal = "<nil>"
  } else {
    successVal = fmt.Sprintf("%v", p.Success)
  }
  return fmt.Sprintf("GeneralStorageServiceRemoveResult({Success:%s})", successVal)
}


