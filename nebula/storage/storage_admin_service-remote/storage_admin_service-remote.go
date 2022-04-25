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
        "github.com/vesoft-inc/nebula-go/v3/nebula/storage"
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
  fmt.Fprintln(os.Stderr, "  DropCPResp dropCheckpoint(DropCPRequest req)")
  fmt.Fprintln(os.Stderr, "  BlockingSignResp blockingWrites(BlockingSignRequest req)")
  fmt.Fprintln(os.Stderr, "  GetLeaderPartsResp getLeaderParts(GetLeaderReq req)")
  fmt.Fprintln(os.Stderr, "  AdminExecResp checkPeers(CheckPeersReq req)")
  fmt.Fprintln(os.Stderr, "  AddTaskResp addAdminTask(AddTaskRequest req)")
  fmt.Fprintln(os.Stderr, "  StopTaskResp stopAdminTask(StopTaskRequest req)")
  fmt.Fprintln(os.Stderr, "  ClearSpaceResp clearSpace(ClearSpaceReq req)")
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
    arg221 := flag.Arg(1)
    mbTrans222 := thrift.NewMemoryBufferLen(len(arg221))
    defer mbTrans222.Close()
    _, err223 := mbTrans222.WriteString(arg221)
    if err223 != nil {
      Usage()
      return
    }
    factory224 := thrift.NewSimpleJSONProtocolFactory()
    jsProt225 := factory224.GetProtocol(mbTrans222)
    argvalue0 := storage.NewTransLeaderReq()
    err226 := argvalue0.Read(jsProt225)
    if err226 != nil {
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
    arg227 := flag.Arg(1)
    mbTrans228 := thrift.NewMemoryBufferLen(len(arg227))
    defer mbTrans228.Close()
    _, err229 := mbTrans228.WriteString(arg227)
    if err229 != nil {
      Usage()
      return
    }
    factory230 := thrift.NewSimpleJSONProtocolFactory()
    jsProt231 := factory230.GetProtocol(mbTrans228)
    argvalue0 := storage.NewAddPartReq()
    err232 := argvalue0.Read(jsProt231)
    if err232 != nil {
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
    arg233 := flag.Arg(1)
    mbTrans234 := thrift.NewMemoryBufferLen(len(arg233))
    defer mbTrans234.Close()
    _, err235 := mbTrans234.WriteString(arg233)
    if err235 != nil {
      Usage()
      return
    }
    factory236 := thrift.NewSimpleJSONProtocolFactory()
    jsProt237 := factory236.GetProtocol(mbTrans234)
    argvalue0 := storage.NewAddLearnerReq()
    err238 := argvalue0.Read(jsProt237)
    if err238 != nil {
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
    arg239 := flag.Arg(1)
    mbTrans240 := thrift.NewMemoryBufferLen(len(arg239))
    defer mbTrans240.Close()
    _, err241 := mbTrans240.WriteString(arg239)
    if err241 != nil {
      Usage()
      return
    }
    factory242 := thrift.NewSimpleJSONProtocolFactory()
    jsProt243 := factory242.GetProtocol(mbTrans240)
    argvalue0 := storage.NewRemovePartReq()
    err244 := argvalue0.Read(jsProt243)
    if err244 != nil {
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
    arg245 := flag.Arg(1)
    mbTrans246 := thrift.NewMemoryBufferLen(len(arg245))
    defer mbTrans246.Close()
    _, err247 := mbTrans246.WriteString(arg245)
    if err247 != nil {
      Usage()
      return
    }
    factory248 := thrift.NewSimpleJSONProtocolFactory()
    jsProt249 := factory248.GetProtocol(mbTrans246)
    argvalue0 := storage.NewMemberChangeReq()
    err250 := argvalue0.Read(jsProt249)
    if err250 != nil {
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
    arg251 := flag.Arg(1)
    mbTrans252 := thrift.NewMemoryBufferLen(len(arg251))
    defer mbTrans252.Close()
    _, err253 := mbTrans252.WriteString(arg251)
    if err253 != nil {
      Usage()
      return
    }
    factory254 := thrift.NewSimpleJSONProtocolFactory()
    jsProt255 := factory254.GetProtocol(mbTrans252)
    argvalue0 := storage.NewCatchUpDataReq()
    err256 := argvalue0.Read(jsProt255)
    if err256 != nil {
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
    arg257 := flag.Arg(1)
    mbTrans258 := thrift.NewMemoryBufferLen(len(arg257))
    defer mbTrans258.Close()
    _, err259 := mbTrans258.WriteString(arg257)
    if err259 != nil {
      Usage()
      return
    }
    factory260 := thrift.NewSimpleJSONProtocolFactory()
    jsProt261 := factory260.GetProtocol(mbTrans258)
    argvalue0 := storage.NewCreateCPRequest()
    err262 := argvalue0.Read(jsProt261)
    if err262 != nil {
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
    arg263 := flag.Arg(1)
    mbTrans264 := thrift.NewMemoryBufferLen(len(arg263))
    defer mbTrans264.Close()
    _, err265 := mbTrans264.WriteString(arg263)
    if err265 != nil {
      Usage()
      return
    }
    factory266 := thrift.NewSimpleJSONProtocolFactory()
    jsProt267 := factory266.GetProtocol(mbTrans264)
    argvalue0 := storage.NewDropCPRequest()
    err268 := argvalue0.Read(jsProt267)
    if err268 != nil {
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
    arg269 := flag.Arg(1)
    mbTrans270 := thrift.NewMemoryBufferLen(len(arg269))
    defer mbTrans270.Close()
    _, err271 := mbTrans270.WriteString(arg269)
    if err271 != nil {
      Usage()
      return
    }
    factory272 := thrift.NewSimpleJSONProtocolFactory()
    jsProt273 := factory272.GetProtocol(mbTrans270)
    argvalue0 := storage.NewBlockingSignRequest()
    err274 := argvalue0.Read(jsProt273)
    if err274 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.BlockingWrites(value0))
    fmt.Print("\n")
    break
  case "getLeaderParts":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "GetLeaderParts requires 1 args")
      flag.Usage()
    }
    arg275 := flag.Arg(1)
    mbTrans276 := thrift.NewMemoryBufferLen(len(arg275))
    defer mbTrans276.Close()
    _, err277 := mbTrans276.WriteString(arg275)
    if err277 != nil {
      Usage()
      return
    }
    factory278 := thrift.NewSimpleJSONProtocolFactory()
    jsProt279 := factory278.GetProtocol(mbTrans276)
    argvalue0 := storage.NewGetLeaderReq()
    err280 := argvalue0.Read(jsProt279)
    if err280 != nil {
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
    arg281 := flag.Arg(1)
    mbTrans282 := thrift.NewMemoryBufferLen(len(arg281))
    defer mbTrans282.Close()
    _, err283 := mbTrans282.WriteString(arg281)
    if err283 != nil {
      Usage()
      return
    }
    factory284 := thrift.NewSimpleJSONProtocolFactory()
    jsProt285 := factory284.GetProtocol(mbTrans282)
    argvalue0 := storage.NewCheckPeersReq()
    err286 := argvalue0.Read(jsProt285)
    if err286 != nil {
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
    arg287 := flag.Arg(1)
    mbTrans288 := thrift.NewMemoryBufferLen(len(arg287))
    defer mbTrans288.Close()
    _, err289 := mbTrans288.WriteString(arg287)
    if err289 != nil {
      Usage()
      return
    }
    factory290 := thrift.NewSimpleJSONProtocolFactory()
    jsProt291 := factory290.GetProtocol(mbTrans288)
    argvalue0 := storage.NewAddTaskRequest()
    err292 := argvalue0.Read(jsProt291)
    if err292 != nil {
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
    arg293 := flag.Arg(1)
    mbTrans294 := thrift.NewMemoryBufferLen(len(arg293))
    defer mbTrans294.Close()
    _, err295 := mbTrans294.WriteString(arg293)
    if err295 != nil {
      Usage()
      return
    }
    factory296 := thrift.NewSimpleJSONProtocolFactory()
    jsProt297 := factory296.GetProtocol(mbTrans294)
    argvalue0 := storage.NewStopTaskRequest()
    err298 := argvalue0.Read(jsProt297)
    if err298 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.StopAdminTask(value0))
    fmt.Print("\n")
    break
  case "clearSpace":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ClearSpace requires 1 args")
      flag.Usage()
    }
    arg299 := flag.Arg(1)
    mbTrans300 := thrift.NewMemoryBufferLen(len(arg299))
    defer mbTrans300.Close()
    _, err301 := mbTrans300.WriteString(arg299)
    if err301 != nil {
      Usage()
      return
    }
    factory302 := thrift.NewSimpleJSONProtocolFactory()
    jsProt303 := factory302.GetProtocol(mbTrans300)
    argvalue0 := storage.NewClearSpaceReq()
    err304 := argvalue0.Read(jsProt303)
    if err304 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ClearSpace(value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
