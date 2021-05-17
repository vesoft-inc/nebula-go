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
        "../../github.com/vesoft-inc/nebula-go/nebula/storage"
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
    arg160 := flag.Arg(1)
    mbTrans161 := thrift.NewMemoryBufferLen(len(arg160))
    defer mbTrans161.Close()
    _, err162 := mbTrans161.WriteString(arg160)
    if err162 != nil {
      Usage()
      return
    }
    factory163 := thrift.NewSimpleJSONProtocolFactory()
    jsProt164 := factory163.GetProtocol(mbTrans161)
    argvalue0 := storage.NewTransLeaderReq()
    err165 := argvalue0.Read(jsProt164)
    if err165 != nil {
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
    arg166 := flag.Arg(1)
    mbTrans167 := thrift.NewMemoryBufferLen(len(arg166))
    defer mbTrans167.Close()
    _, err168 := mbTrans167.WriteString(arg166)
    if err168 != nil {
      Usage()
      return
    }
    factory169 := thrift.NewSimpleJSONProtocolFactory()
    jsProt170 := factory169.GetProtocol(mbTrans167)
    argvalue0 := storage.NewAddPartReq()
    err171 := argvalue0.Read(jsProt170)
    if err171 != nil {
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
    arg172 := flag.Arg(1)
    mbTrans173 := thrift.NewMemoryBufferLen(len(arg172))
    defer mbTrans173.Close()
    _, err174 := mbTrans173.WriteString(arg172)
    if err174 != nil {
      Usage()
      return
    }
    factory175 := thrift.NewSimpleJSONProtocolFactory()
    jsProt176 := factory175.GetProtocol(mbTrans173)
    argvalue0 := storage.NewAddLearnerReq()
    err177 := argvalue0.Read(jsProt176)
    if err177 != nil {
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
    arg178 := flag.Arg(1)
    mbTrans179 := thrift.NewMemoryBufferLen(len(arg178))
    defer mbTrans179.Close()
    _, err180 := mbTrans179.WriteString(arg178)
    if err180 != nil {
      Usage()
      return
    }
    factory181 := thrift.NewSimpleJSONProtocolFactory()
    jsProt182 := factory181.GetProtocol(mbTrans179)
    argvalue0 := storage.NewRemovePartReq()
    err183 := argvalue0.Read(jsProt182)
    if err183 != nil {
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
    arg184 := flag.Arg(1)
    mbTrans185 := thrift.NewMemoryBufferLen(len(arg184))
    defer mbTrans185.Close()
    _, err186 := mbTrans185.WriteString(arg184)
    if err186 != nil {
      Usage()
      return
    }
    factory187 := thrift.NewSimpleJSONProtocolFactory()
    jsProt188 := factory187.GetProtocol(mbTrans185)
    argvalue0 := storage.NewMemberChangeReq()
    err189 := argvalue0.Read(jsProt188)
    if err189 != nil {
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
    arg190 := flag.Arg(1)
    mbTrans191 := thrift.NewMemoryBufferLen(len(arg190))
    defer mbTrans191.Close()
    _, err192 := mbTrans191.WriteString(arg190)
    if err192 != nil {
      Usage()
      return
    }
    factory193 := thrift.NewSimpleJSONProtocolFactory()
    jsProt194 := factory193.GetProtocol(mbTrans191)
    argvalue0 := storage.NewCatchUpDataReq()
    err195 := argvalue0.Read(jsProt194)
    if err195 != nil {
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
    arg196 := flag.Arg(1)
    mbTrans197 := thrift.NewMemoryBufferLen(len(arg196))
    defer mbTrans197.Close()
    _, err198 := mbTrans197.WriteString(arg196)
    if err198 != nil {
      Usage()
      return
    }
    factory199 := thrift.NewSimpleJSONProtocolFactory()
    jsProt200 := factory199.GetProtocol(mbTrans197)
    argvalue0 := storage.NewCreateCPRequest()
    err201 := argvalue0.Read(jsProt200)
    if err201 != nil {
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
    arg202 := flag.Arg(1)
    mbTrans203 := thrift.NewMemoryBufferLen(len(arg202))
    defer mbTrans203.Close()
    _, err204 := mbTrans203.WriteString(arg202)
    if err204 != nil {
      Usage()
      return
    }
    factory205 := thrift.NewSimpleJSONProtocolFactory()
    jsProt206 := factory205.GetProtocol(mbTrans203)
    argvalue0 := storage.NewDropCPRequest()
    err207 := argvalue0.Read(jsProt206)
    if err207 != nil {
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
    arg208 := flag.Arg(1)
    mbTrans209 := thrift.NewMemoryBufferLen(len(arg208))
    defer mbTrans209.Close()
    _, err210 := mbTrans209.WriteString(arg208)
    if err210 != nil {
      Usage()
      return
    }
    factory211 := thrift.NewSimpleJSONProtocolFactory()
    jsProt212 := factory211.GetProtocol(mbTrans209)
    argvalue0 := storage.NewBlockingSignRequest()
    err213 := argvalue0.Read(jsProt212)
    if err213 != nil {
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
    arg214 := flag.Arg(1)
    mbTrans215 := thrift.NewMemoryBufferLen(len(arg214))
    defer mbTrans215.Close()
    _, err216 := mbTrans215.WriteString(arg214)
    if err216 != nil {
      Usage()
      return
    }
    factory217 := thrift.NewSimpleJSONProtocolFactory()
    jsProt218 := factory217.GetProtocol(mbTrans215)
    argvalue0 := storage.NewRebuildIndexRequest()
    err219 := argvalue0.Read(jsProt218)
    if err219 != nil {
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
    arg220 := flag.Arg(1)
    mbTrans221 := thrift.NewMemoryBufferLen(len(arg220))
    defer mbTrans221.Close()
    _, err222 := mbTrans221.WriteString(arg220)
    if err222 != nil {
      Usage()
      return
    }
    factory223 := thrift.NewSimpleJSONProtocolFactory()
    jsProt224 := factory223.GetProtocol(mbTrans221)
    argvalue0 := storage.NewRebuildIndexRequest()
    err225 := argvalue0.Read(jsProt224)
    if err225 != nil {
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
    arg226 := flag.Arg(1)
    mbTrans227 := thrift.NewMemoryBufferLen(len(arg226))
    defer mbTrans227.Close()
    _, err228 := mbTrans227.WriteString(arg226)
    if err228 != nil {
      Usage()
      return
    }
    factory229 := thrift.NewSimpleJSONProtocolFactory()
    jsProt230 := factory229.GetProtocol(mbTrans227)
    argvalue0 := storage.NewGetLeaderReq()
    err231 := argvalue0.Read(jsProt230)
    if err231 != nil {
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
    arg232 := flag.Arg(1)
    mbTrans233 := thrift.NewMemoryBufferLen(len(arg232))
    defer mbTrans233.Close()
    _, err234 := mbTrans233.WriteString(arg232)
    if err234 != nil {
      Usage()
      return
    }
    factory235 := thrift.NewSimpleJSONProtocolFactory()
    jsProt236 := factory235.GetProtocol(mbTrans233)
    argvalue0 := storage.NewCheckPeersReq()
    err237 := argvalue0.Read(jsProt236)
    if err237 != nil {
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
    arg238 := flag.Arg(1)
    mbTrans239 := thrift.NewMemoryBufferLen(len(arg238))
    defer mbTrans239.Close()
    _, err240 := mbTrans239.WriteString(arg238)
    if err240 != nil {
      Usage()
      return
    }
    factory241 := thrift.NewSimpleJSONProtocolFactory()
    jsProt242 := factory241.GetProtocol(mbTrans239)
    argvalue0 := storage.NewAddAdminTaskRequest()
    err243 := argvalue0.Read(jsProt242)
    if err243 != nil {
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
    arg244 := flag.Arg(1)
    mbTrans245 := thrift.NewMemoryBufferLen(len(arg244))
    defer mbTrans245.Close()
    _, err246 := mbTrans245.WriteString(arg244)
    if err246 != nil {
      Usage()
      return
    }
    factory247 := thrift.NewSimpleJSONProtocolFactory()
    jsProt248 := factory247.GetProtocol(mbTrans245)
    argvalue0 := storage.NewStopAdminTaskRequest()
    err249 := argvalue0.Read(jsProt248)
    if err249 != nil {
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
    arg250 := flag.Arg(1)
    mbTrans251 := thrift.NewMemoryBufferLen(len(arg250))
    defer mbTrans251.Close()
    _, err252 := mbTrans251.WriteString(arg250)
    if err252 != nil {
      Usage()
      return
    }
    factory253 := thrift.NewSimpleJSONProtocolFactory()
    jsProt254 := factory253.GetProtocol(mbTrans251)
    argvalue0 := storage.NewListClusterInfoReq()
    err255 := argvalue0.Read(jsProt254)
    if err255 != nil {
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
