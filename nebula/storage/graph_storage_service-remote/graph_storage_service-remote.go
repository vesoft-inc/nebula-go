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
  fmt.Fprintln(os.Stderr, "  GetNeighborsResponse getNeighbors(GetNeighborsRequest req)")
  fmt.Fprintln(os.Stderr, "  GetPropResponse getProps(GetPropRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse addVertices(AddVerticesRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse addEdges(AddEdgesRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse deleteEdges(DeleteEdgesRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse deleteVertices(DeleteVerticesRequest req)")
  fmt.Fprintln(os.Stderr, "  UpdateResponse updateVertex(UpdateVertexRequest req)")
  fmt.Fprintln(os.Stderr, "  UpdateResponse updateEdge(UpdateEdgeRequest req)")
  fmt.Fprintln(os.Stderr, "  ScanVertexResponse scanVertex(ScanVertexRequest req)")
  fmt.Fprintln(os.Stderr, "  ScanEdgeResponse scanEdge(ScanEdgeRequest req)")
  fmt.Fprintln(os.Stderr, "  GetUUIDResp getUUID(GetUUIDReq req)")
  fmt.Fprintln(os.Stderr, "  LookupIndexResp lookupIndex(LookupIndexRequest req)")
  fmt.Fprintln(os.Stderr, "  GetNeighborsResponse lookupAndTraverse(LookupAndTraverseRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse addEdgesAtomic(AddEdgesRequest req)")
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
  client := storage.NewGraphStorageServiceClientFactory(trans, protocolFactory)
  if err := trans.Open(); err != nil {
    fmt.Fprintln(os.Stderr, "Error opening socket to ", host, ":", port, " ", err)
    os.Exit(1)
  }
  
  switch cmd {
  case "getNeighbors":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "GetNeighbors requires 1 args")
      flag.Usage()
    }
    arg73 := flag.Arg(1)
    mbTrans74 := thrift.NewMemoryBufferLen(len(arg73))
    defer mbTrans74.Close()
    _, err75 := mbTrans74.WriteString(arg73)
    if err75 != nil {
      Usage()
      return
    }
    factory76 := thrift.NewSimpleJSONProtocolFactory()
    jsProt77 := factory76.GetProtocol(mbTrans74)
    argvalue0 := storage.NewGetNeighborsRequest()
    err78 := argvalue0.Read(jsProt77)
    if err78 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.GetNeighbors(value0))
    fmt.Print("\n")
    break
  case "getProps":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "GetProps requires 1 args")
      flag.Usage()
    }
    arg79 := flag.Arg(1)
    mbTrans80 := thrift.NewMemoryBufferLen(len(arg79))
    defer mbTrans80.Close()
    _, err81 := mbTrans80.WriteString(arg79)
    if err81 != nil {
      Usage()
      return
    }
    factory82 := thrift.NewSimpleJSONProtocolFactory()
    jsProt83 := factory82.GetProtocol(mbTrans80)
    argvalue0 := storage.NewGetPropRequest()
    err84 := argvalue0.Read(jsProt83)
    if err84 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.GetProps(value0))
    fmt.Print("\n")
    break
  case "addVertices":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "AddVertices requires 1 args")
      flag.Usage()
    }
    arg85 := flag.Arg(1)
    mbTrans86 := thrift.NewMemoryBufferLen(len(arg85))
    defer mbTrans86.Close()
    _, err87 := mbTrans86.WriteString(arg85)
    if err87 != nil {
      Usage()
      return
    }
    factory88 := thrift.NewSimpleJSONProtocolFactory()
    jsProt89 := factory88.GetProtocol(mbTrans86)
    argvalue0 := storage.NewAddVerticesRequest()
    err90 := argvalue0.Read(jsProt89)
    if err90 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.AddVertices(value0))
    fmt.Print("\n")
    break
  case "addEdges":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "AddEdges requires 1 args")
      flag.Usage()
    }
    arg91 := flag.Arg(1)
    mbTrans92 := thrift.NewMemoryBufferLen(len(arg91))
    defer mbTrans92.Close()
    _, err93 := mbTrans92.WriteString(arg91)
    if err93 != nil {
      Usage()
      return
    }
    factory94 := thrift.NewSimpleJSONProtocolFactory()
    jsProt95 := factory94.GetProtocol(mbTrans92)
    argvalue0 := storage.NewAddEdgesRequest()
    err96 := argvalue0.Read(jsProt95)
    if err96 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.AddEdges(value0))
    fmt.Print("\n")
    break
  case "deleteEdges":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "DeleteEdges requires 1 args")
      flag.Usage()
    }
    arg97 := flag.Arg(1)
    mbTrans98 := thrift.NewMemoryBufferLen(len(arg97))
    defer mbTrans98.Close()
    _, err99 := mbTrans98.WriteString(arg97)
    if err99 != nil {
      Usage()
      return
    }
    factory100 := thrift.NewSimpleJSONProtocolFactory()
    jsProt101 := factory100.GetProtocol(mbTrans98)
    argvalue0 := storage.NewDeleteEdgesRequest()
    err102 := argvalue0.Read(jsProt101)
    if err102 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.DeleteEdges(value0))
    fmt.Print("\n")
    break
  case "deleteVertices":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "DeleteVertices requires 1 args")
      flag.Usage()
    }
    arg103 := flag.Arg(1)
    mbTrans104 := thrift.NewMemoryBufferLen(len(arg103))
    defer mbTrans104.Close()
    _, err105 := mbTrans104.WriteString(arg103)
    if err105 != nil {
      Usage()
      return
    }
    factory106 := thrift.NewSimpleJSONProtocolFactory()
    jsProt107 := factory106.GetProtocol(mbTrans104)
    argvalue0 := storage.NewDeleteVerticesRequest()
    err108 := argvalue0.Read(jsProt107)
    if err108 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.DeleteVertices(value0))
    fmt.Print("\n")
    break
  case "updateVertex":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "UpdateVertex requires 1 args")
      flag.Usage()
    }
    arg109 := flag.Arg(1)
    mbTrans110 := thrift.NewMemoryBufferLen(len(arg109))
    defer mbTrans110.Close()
    _, err111 := mbTrans110.WriteString(arg109)
    if err111 != nil {
      Usage()
      return
    }
    factory112 := thrift.NewSimpleJSONProtocolFactory()
    jsProt113 := factory112.GetProtocol(mbTrans110)
    argvalue0 := storage.NewUpdateVertexRequest()
    err114 := argvalue0.Read(jsProt113)
    if err114 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.UpdateVertex(value0))
    fmt.Print("\n")
    break
  case "updateEdge":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "UpdateEdge requires 1 args")
      flag.Usage()
    }
    arg115 := flag.Arg(1)
    mbTrans116 := thrift.NewMemoryBufferLen(len(arg115))
    defer mbTrans116.Close()
    _, err117 := mbTrans116.WriteString(arg115)
    if err117 != nil {
      Usage()
      return
    }
    factory118 := thrift.NewSimpleJSONProtocolFactory()
    jsProt119 := factory118.GetProtocol(mbTrans116)
    argvalue0 := storage.NewUpdateEdgeRequest()
    err120 := argvalue0.Read(jsProt119)
    if err120 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.UpdateEdge(value0))
    fmt.Print("\n")
    break
  case "scanVertex":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ScanVertex requires 1 args")
      flag.Usage()
    }
    arg121 := flag.Arg(1)
    mbTrans122 := thrift.NewMemoryBufferLen(len(arg121))
    defer mbTrans122.Close()
    _, err123 := mbTrans122.WriteString(arg121)
    if err123 != nil {
      Usage()
      return
    }
    factory124 := thrift.NewSimpleJSONProtocolFactory()
    jsProt125 := factory124.GetProtocol(mbTrans122)
    argvalue0 := storage.NewScanVertexRequest()
    err126 := argvalue0.Read(jsProt125)
    if err126 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ScanVertex(value0))
    fmt.Print("\n")
    break
  case "scanEdge":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ScanEdge requires 1 args")
      flag.Usage()
    }
    arg127 := flag.Arg(1)
    mbTrans128 := thrift.NewMemoryBufferLen(len(arg127))
    defer mbTrans128.Close()
    _, err129 := mbTrans128.WriteString(arg127)
    if err129 != nil {
      Usage()
      return
    }
    factory130 := thrift.NewSimpleJSONProtocolFactory()
    jsProt131 := factory130.GetProtocol(mbTrans128)
    argvalue0 := storage.NewScanEdgeRequest()
    err132 := argvalue0.Read(jsProt131)
    if err132 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ScanEdge(value0))
    fmt.Print("\n")
    break
  case "getUUID":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "GetUUID requires 1 args")
      flag.Usage()
    }
    arg133 := flag.Arg(1)
    mbTrans134 := thrift.NewMemoryBufferLen(len(arg133))
    defer mbTrans134.Close()
    _, err135 := mbTrans134.WriteString(arg133)
    if err135 != nil {
      Usage()
      return
    }
    factory136 := thrift.NewSimpleJSONProtocolFactory()
    jsProt137 := factory136.GetProtocol(mbTrans134)
    argvalue0 := storage.NewGetUUIDReq()
    err138 := argvalue0.Read(jsProt137)
    if err138 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.GetUUID(value0))
    fmt.Print("\n")
    break
  case "lookupIndex":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "LookupIndex requires 1 args")
      flag.Usage()
    }
    arg139 := flag.Arg(1)
    mbTrans140 := thrift.NewMemoryBufferLen(len(arg139))
    defer mbTrans140.Close()
    _, err141 := mbTrans140.WriteString(arg139)
    if err141 != nil {
      Usage()
      return
    }
    factory142 := thrift.NewSimpleJSONProtocolFactory()
    jsProt143 := factory142.GetProtocol(mbTrans140)
    argvalue0 := storage.NewLookupIndexRequest()
    err144 := argvalue0.Read(jsProt143)
    if err144 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.LookupIndex(value0))
    fmt.Print("\n")
    break
  case "lookupAndTraverse":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "LookupAndTraverse requires 1 args")
      flag.Usage()
    }
    arg145 := flag.Arg(1)
    mbTrans146 := thrift.NewMemoryBufferLen(len(arg145))
    defer mbTrans146.Close()
    _, err147 := mbTrans146.WriteString(arg145)
    if err147 != nil {
      Usage()
      return
    }
    factory148 := thrift.NewSimpleJSONProtocolFactory()
    jsProt149 := factory148.GetProtocol(mbTrans146)
    argvalue0 := storage.NewLookupAndTraverseRequest()
    err150 := argvalue0.Read(jsProt149)
    if err150 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.LookupAndTraverse(value0))
    fmt.Print("\n")
    break
  case "addEdgesAtomic":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "AddEdgesAtomic requires 1 args")
      flag.Usage()
    }
    arg151 := flag.Arg(1)
    mbTrans152 := thrift.NewMemoryBufferLen(len(arg151))
    defer mbTrans152.Close()
    _, err153 := mbTrans152.WriteString(arg151)
    if err153 != nil {
      Usage()
      return
    }
    factory154 := thrift.NewSimpleJSONProtocolFactory()
    jsProt155 := factory154.GetProtocol(mbTrans152)
    argvalue0 := storage.NewAddEdgesRequest()
    err156 := argvalue0.Read(jsProt155)
    if err156 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.AddEdgesAtomic(value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
