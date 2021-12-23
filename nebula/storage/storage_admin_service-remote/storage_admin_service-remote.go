// Autogenerated by Thrift Compiler (facebook)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
// @generated

package main

import (
        "flag"
        "fmt"
        "math"
        "net"
        "net/url"
        "os"
        "strconv"
        "strings"
        thrift "github.com/facebook/fbthrift/thrift/lib/go/thrift"
        "../../github.com/vesoft-inc/nebula-go/v2/nebula/storage"
)

func Usage() {
  fmt.Fprintln(os.Stderr, "Usage of ", os.Args[0], " [-h host:port] [-u url] [-f[ramed]] function [arg1 [arg2...]]:")
  flag.PrintDefaults()
  fmt.Fprintln(os.Stderr, "\nFunctions:")
  fmt.Fprintln(os.Stderr, "  AdminExecResp transLeader(TransLeaderReq req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp addPart(AddPartReq req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp addLearner(AddLearnerReq req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp removePart(RemovePartReq req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp memberChange(MemberChangeReq req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp waitingForCatchUpData(CatchUpDataReq req)")
  fmt.Fprintln(os.Stderr, "  CreateCPResp createCheckpoint(CreateCPRequest req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp dropCheckpoint(DropCPRequest req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp blockingWrites(BlockingSignRequest req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp rebuildTagIndex(RebuildIndexRequest req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp rebuildEdgeIndex(RebuildIndexRequest req)")
  fmt.Fprintln(os.Stderr, "  GetLeaderPartsResp getLeaderParts(GetLeaderReq req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp checkPeers(CheckPeersReq req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp addAdminTask(AddAdminTaskRequest req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp stopAdminTask(StopAdminTaskRequest req)")
  fmt.Fprintln(os.Stderr, "  ListClusterInfoResp listClusterInfo(ListClusterInfoReq req)")
  fmt.Fprintln(os.Stderr)
  os.Exit(0)
}

func main() {
  flag.Usage = Usage
  var host string
  var port int
  var protocol string
  var urlString string
  var framed bool
  var useHttp bool
  var parsedUrl url.URL
  var trans thrift.Transport
  _ = strconv.Atoi
  _ = math.Abs
  flag.Usage = Usage
  flag.StringVar(&host, "h", "localhost", "Specify host")
  flag.IntVar(&port, "p", 9090, "Specify port")
  flag.StringVar(&protocol, "P", "binary", "Specify the protocol (binary, compact, simplejson, json)")
  flag.StringVar(&urlString, "u", "", "Specify the url")
  flag.BoolVar(&framed, "framed", false, "Use framed transport")
  flag.BoolVar(&useHttp, "http", false, "Use http")
  flag.Parse()
  
  if len(urlString) > 0 {
    parsedUrl, err := url.Parse(urlString)
    if err != nil {
      fmt.Fprintln(os.Stderr, "Error parsing URL: ", err)
      flag.Usage()
    }
    host = parsedUrl.Host
    useHttp = len(parsedUrl.Scheme) <= 0 || parsedUrl.Scheme == "http"
  } else if useHttp {
    _, err := url.Parse(fmt.Sprint("http://", host, ":", port))
    if err != nil {
      fmt.Fprintln(os.Stderr, "Error parsing URL: ", err)
      flag.Usage()
    }
  }
  
  cmd := flag.Arg(0)
  var err error
  if useHttp {
    trans, err = thrift.NewHTTPPostClient(parsedUrl.String())
  } else {
    portStr := fmt.Sprint(port)
    if strings.Contains(host, ":") {
           host, portStr, err = net.SplitHostPort(host)
           if err != nil {
                   fmt.Fprintln(os.Stderr, "error with host:", err)
                   os.Exit(1)
           }
    }
    trans, err = thrift.NewSocket(thrift.SocketAddr(net.JoinHostPort(host, portStr)))
    if err != nil {
      fmt.Fprintln(os.Stderr, "error resolving address:", err)
      os.Exit(1)
    }
    if framed {
      trans = thrift.NewFramedTransport(trans)
    }
  }
  if err != nil {
    fmt.Fprintln(os.Stderr, "Error creating transport", err)
    os.Exit(1)
  }
  defer trans.Close()
  var protocolFactory thrift.ProtocolFactory
  switch protocol {
  case "compact":
    protocolFactory = thrift.NewCompactProtocolFactory()
    break
  case "simplejson":
    protocolFactory = thrift.NewSimpleJSONProtocolFactory()
    break
  case "json":
    protocolFactory = thrift.NewJSONProtocolFactory()
    break
  case "binary", "":
    protocolFactory = thrift.NewBinaryProtocolFactoryDefault()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid protocol specified: ", protocol)
    Usage()
    os.Exit(1)
  }
  client := storage.NewStorageAdminServiceClientFactory(trans, protocolFactory)
  if err := trans.Open(); err != nil {
    fmt.Fprintln(os.Stderr, "Error opening socket to ", host, ":", port, " ", err)
    os.Exit(1)
  }
  
  switch cmd {
  case "transLeader":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "TransLeader requires 1 args")
      flag.Usage()
    }
    arg212 := flag.Arg(1)
    mbTrans213 := thrift.NewMemoryBufferLen(len(arg212))
    defer mbTrans213.Close()
    _, err214 := mbTrans213.WriteString(arg212)
    if err214 != nil {
      Usage()
      return
    }
    factory215 := thrift.NewSimpleJSONProtocolFactory()
    jsProt216 := factory215.GetProtocol(mbTrans213)
    argvalue0 := storage.NewTransLeaderReq()
    err217 := argvalue0.Read(jsProt216)
    if err217 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.TransLeader(value0))
    fmt.Print("\n")
    break
  case "addPart":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "AddPart requires 1 args")
      flag.Usage()
    }
    arg218 := flag.Arg(1)
    mbTrans219 := thrift.NewMemoryBufferLen(len(arg218))
    defer mbTrans219.Close()
    _, err220 := mbTrans219.WriteString(arg218)
    if err220 != nil {
      Usage()
      return
    }
    factory221 := thrift.NewSimpleJSONProtocolFactory()
    jsProt222 := factory221.GetProtocol(mbTrans219)
    argvalue0 := storage.NewAddPartReq()
    err223 := argvalue0.Read(jsProt222)
    if err223 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.AddPart(value0))
    fmt.Print("\n")
    break
  case "addLearner":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "AddLearner requires 1 args")
      flag.Usage()
    }
    arg224 := flag.Arg(1)
    mbTrans225 := thrift.NewMemoryBufferLen(len(arg224))
    defer mbTrans225.Close()
    _, err226 := mbTrans225.WriteString(arg224)
    if err226 != nil {
      Usage()
      return
    }
    factory227 := thrift.NewSimpleJSONProtocolFactory()
    jsProt228 := factory227.GetProtocol(mbTrans225)
    argvalue0 := storage.NewAddLearnerReq()
    err229 := argvalue0.Read(jsProt228)
    if err229 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.AddLearner(value0))
    fmt.Print("\n")
    break
  case "removePart":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "RemovePart requires 1 args")
      flag.Usage()
    }
    arg230 := flag.Arg(1)
    mbTrans231 := thrift.NewMemoryBufferLen(len(arg230))
    defer mbTrans231.Close()
    _, err232 := mbTrans231.WriteString(arg230)
    if err232 != nil {
      Usage()
      return
    }
    factory233 := thrift.NewSimpleJSONProtocolFactory()
    jsProt234 := factory233.GetProtocol(mbTrans231)
    argvalue0 := storage.NewRemovePartReq()
    err235 := argvalue0.Read(jsProt234)
    if err235 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.RemovePart(value0))
    fmt.Print("\n")
    break
  case "memberChange":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "MemberChange requires 1 args")
      flag.Usage()
    }
    arg236 := flag.Arg(1)
    mbTrans237 := thrift.NewMemoryBufferLen(len(arg236))
    defer mbTrans237.Close()
    _, err238 := mbTrans237.WriteString(arg236)
    if err238 != nil {
      Usage()
      return
    }
    factory239 := thrift.NewSimpleJSONProtocolFactory()
    jsProt240 := factory239.GetProtocol(mbTrans237)
    argvalue0 := storage.NewMemberChangeReq()
    err241 := argvalue0.Read(jsProt240)
    if err241 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.MemberChange(value0))
    fmt.Print("\n")
    break
  case "waitingForCatchUpData":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "WaitingForCatchUpData requires 1 args")
      flag.Usage()
    }
    arg242 := flag.Arg(1)
    mbTrans243 := thrift.NewMemoryBufferLen(len(arg242))
    defer mbTrans243.Close()
    _, err244 := mbTrans243.WriteString(arg242)
    if err244 != nil {
      Usage()
      return
    }
    factory245 := thrift.NewSimpleJSONProtocolFactory()
    jsProt246 := factory245.GetProtocol(mbTrans243)
    argvalue0 := storage.NewCatchUpDataReq()
    err247 := argvalue0.Read(jsProt246)
    if err247 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.WaitingForCatchUpData(value0))
    fmt.Print("\n")
    break
  case "createCheckpoint":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "CreateCheckpoint requires 1 args")
      flag.Usage()
    }
    arg248 := flag.Arg(1)
    mbTrans249 := thrift.NewMemoryBufferLen(len(arg248))
    defer mbTrans249.Close()
    _, err250 := mbTrans249.WriteString(arg248)
    if err250 != nil {
      Usage()
      return
    }
    factory251 := thrift.NewSimpleJSONProtocolFactory()
    jsProt252 := factory251.GetProtocol(mbTrans249)
    argvalue0 := storage.NewCreateCPRequest()
    err253 := argvalue0.Read(jsProt252)
    if err253 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.CreateCheckpoint(value0))
    fmt.Print("\n")
    break
  case "dropCheckpoint":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "DropCheckpoint requires 1 args")
      flag.Usage()
    }
    arg254 := flag.Arg(1)
    mbTrans255 := thrift.NewMemoryBufferLen(len(arg254))
    defer mbTrans255.Close()
    _, err256 := mbTrans255.WriteString(arg254)
    if err256 != nil {
      Usage()
      return
    }
    factory257 := thrift.NewSimpleJSONProtocolFactory()
    jsProt258 := factory257.GetProtocol(mbTrans255)
    argvalue0 := storage.NewDropCPRequest()
    err259 := argvalue0.Read(jsProt258)
    if err259 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.DropCheckpoint(value0))
    fmt.Print("\n")
    break
  case "blockingWrites":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "BlockingWrites requires 1 args")
      flag.Usage()
    }
    arg260 := flag.Arg(1)
    mbTrans261 := thrift.NewMemoryBufferLen(len(arg260))
    defer mbTrans261.Close()
    _, err262 := mbTrans261.WriteString(arg260)
    if err262 != nil {
      Usage()
      return
    }
    factory263 := thrift.NewSimpleJSONProtocolFactory()
    jsProt264 := factory263.GetProtocol(mbTrans261)
    argvalue0 := storage.NewBlockingSignRequest()
    err265 := argvalue0.Read(jsProt264)
    if err265 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.BlockingWrites(value0))
    fmt.Print("\n")
    break
  case "rebuildTagIndex":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "RebuildTagIndex requires 1 args")
      flag.Usage()
    }
    arg266 := flag.Arg(1)
    mbTrans267 := thrift.NewMemoryBufferLen(len(arg266))
    defer mbTrans267.Close()
    _, err268 := mbTrans267.WriteString(arg266)
    if err268 != nil {
      Usage()
      return
    }
    factory269 := thrift.NewSimpleJSONProtocolFactory()
    jsProt270 := factory269.GetProtocol(mbTrans267)
    argvalue0 := storage.NewRebuildIndexRequest()
    err271 := argvalue0.Read(jsProt270)
    if err271 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.RebuildTagIndex(value0))
    fmt.Print("\n")
    break
  case "rebuildEdgeIndex":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "RebuildEdgeIndex requires 1 args")
      flag.Usage()
    }
    arg272 := flag.Arg(1)
    mbTrans273 := thrift.NewMemoryBufferLen(len(arg272))
    defer mbTrans273.Close()
    _, err274 := mbTrans273.WriteString(arg272)
    if err274 != nil {
      Usage()
      return
    }
    factory275 := thrift.NewSimpleJSONProtocolFactory()
    jsProt276 := factory275.GetProtocol(mbTrans273)
    argvalue0 := storage.NewRebuildIndexRequest()
    err277 := argvalue0.Read(jsProt276)
    if err277 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.RebuildEdgeIndex(value0))
    fmt.Print("\n")
    break
  case "getLeaderParts":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "GetLeaderParts requires 1 args")
      flag.Usage()
    }
    arg278 := flag.Arg(1)
    mbTrans279 := thrift.NewMemoryBufferLen(len(arg278))
    defer mbTrans279.Close()
    _, err280 := mbTrans279.WriteString(arg278)
    if err280 != nil {
      Usage()
      return
    }
    factory281 := thrift.NewSimpleJSONProtocolFactory()
    jsProt282 := factory281.GetProtocol(mbTrans279)
    argvalue0 := storage.NewGetLeaderReq()
    err283 := argvalue0.Read(jsProt282)
    if err283 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.GetLeaderParts(value0))
    fmt.Print("\n")
    break
  case "checkPeers":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "CheckPeers requires 1 args")
      flag.Usage()
    }
    arg284 := flag.Arg(1)
    mbTrans285 := thrift.NewMemoryBufferLen(len(arg284))
    defer mbTrans285.Close()
    _, err286 := mbTrans285.WriteString(arg284)
    if err286 != nil {
      Usage()
      return
    }
    factory287 := thrift.NewSimpleJSONProtocolFactory()
    jsProt288 := factory287.GetProtocol(mbTrans285)
    argvalue0 := storage.NewCheckPeersReq()
    err289 := argvalue0.Read(jsProt288)
    if err289 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.CheckPeers(value0))
    fmt.Print("\n")
    break
  case "addAdminTask":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "AddAdminTask requires 1 args")
      flag.Usage()
    }
    arg290 := flag.Arg(1)
    mbTrans291 := thrift.NewMemoryBufferLen(len(arg290))
    defer mbTrans291.Close()
    _, err292 := mbTrans291.WriteString(arg290)
    if err292 != nil {
      Usage()
      return
    }
    factory293 := thrift.NewSimpleJSONProtocolFactory()
    jsProt294 := factory293.GetProtocol(mbTrans291)
    argvalue0 := storage.NewAddAdminTaskRequest()
    err295 := argvalue0.Read(jsProt294)
    if err295 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.AddAdminTask(value0))
    fmt.Print("\n")
    break
  case "stopAdminTask":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "StopAdminTask requires 1 args")
      flag.Usage()
    }
    arg296 := flag.Arg(1)
    mbTrans297 := thrift.NewMemoryBufferLen(len(arg296))
    defer mbTrans297.Close()
    _, err298 := mbTrans297.WriteString(arg296)
    if err298 != nil {
      Usage()
      return
    }
    factory299 := thrift.NewSimpleJSONProtocolFactory()
    jsProt300 := factory299.GetProtocol(mbTrans297)
    argvalue0 := storage.NewStopAdminTaskRequest()
    err301 := argvalue0.Read(jsProt300)
    if err301 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.StopAdminTask(value0))
    fmt.Print("\n")
    break
  case "listClusterInfo":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ListClusterInfo requires 1 args")
      flag.Usage()
    }
    arg302 := flag.Arg(1)
    mbTrans303 := thrift.NewMemoryBufferLen(len(arg302))
    defer mbTrans303.Close()
    _, err304 := mbTrans303.WriteString(arg302)
    if err304 != nil {
      Usage()
      return
    }
    factory305 := thrift.NewSimpleJSONProtocolFactory()
    jsProt306 := factory305.GetProtocol(mbTrans303)
    argvalue0 := storage.NewListClusterInfoReq()
    err307 := argvalue0.Read(jsProt306)
    if err307 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ListClusterInfo(value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
