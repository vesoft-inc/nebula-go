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
    arg74 := flag.Arg(1)
    mbTrans75 := thrift.NewMemoryBufferLen(len(arg74))
    defer mbTrans75.Close()
    _, err76 := mbTrans75.WriteString(arg74)
    if err76 != nil {
      Usage()
      return
    }
    factory77 := thrift.NewSimpleJSONProtocolFactory()
    jsProt78 := factory77.GetProtocol(mbTrans75)
    argvalue0 := storage.NewGetNeighborsRequest()
    err79 := argvalue0.Read(jsProt78)
    if err79 != nil {
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
    arg80 := flag.Arg(1)
    mbTrans81 := thrift.NewMemoryBufferLen(len(arg80))
    defer mbTrans81.Close()
    _, err82 := mbTrans81.WriteString(arg80)
    if err82 != nil {
      Usage()
      return
    }
    factory83 := thrift.NewSimpleJSONProtocolFactory()
    jsProt84 := factory83.GetProtocol(mbTrans81)
    argvalue0 := storage.NewGetPropRequest()
    err85 := argvalue0.Read(jsProt84)
    if err85 != nil {
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
    arg86 := flag.Arg(1)
    mbTrans87 := thrift.NewMemoryBufferLen(len(arg86))
    defer mbTrans87.Close()
    _, err88 := mbTrans87.WriteString(arg86)
    if err88 != nil {
      Usage()
      return
    }
    factory89 := thrift.NewSimpleJSONProtocolFactory()
    jsProt90 := factory89.GetProtocol(mbTrans87)
    argvalue0 := storage.NewAddVerticesRequest()
    err91 := argvalue0.Read(jsProt90)
    if err91 != nil {
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
    arg92 := flag.Arg(1)
    mbTrans93 := thrift.NewMemoryBufferLen(len(arg92))
    defer mbTrans93.Close()
    _, err94 := mbTrans93.WriteString(arg92)
    if err94 != nil {
      Usage()
      return
    }
    factory95 := thrift.NewSimpleJSONProtocolFactory()
    jsProt96 := factory95.GetProtocol(mbTrans93)
    argvalue0 := storage.NewAddEdgesRequest()
    err97 := argvalue0.Read(jsProt96)
    if err97 != nil {
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
    arg98 := flag.Arg(1)
    mbTrans99 := thrift.NewMemoryBufferLen(len(arg98))
    defer mbTrans99.Close()
    _, err100 := mbTrans99.WriteString(arg98)
    if err100 != nil {
      Usage()
      return
    }
    factory101 := thrift.NewSimpleJSONProtocolFactory()
    jsProt102 := factory101.GetProtocol(mbTrans99)
    argvalue0 := storage.NewDeleteEdgesRequest()
    err103 := argvalue0.Read(jsProt102)
    if err103 != nil {
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
    arg104 := flag.Arg(1)
    mbTrans105 := thrift.NewMemoryBufferLen(len(arg104))
    defer mbTrans105.Close()
    _, err106 := mbTrans105.WriteString(arg104)
    if err106 != nil {
      Usage()
      return
    }
    factory107 := thrift.NewSimpleJSONProtocolFactory()
    jsProt108 := factory107.GetProtocol(mbTrans105)
    argvalue0 := storage.NewDeleteVerticesRequest()
    err109 := argvalue0.Read(jsProt108)
    if err109 != nil {
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
    arg110 := flag.Arg(1)
    mbTrans111 := thrift.NewMemoryBufferLen(len(arg110))
    defer mbTrans111.Close()
    _, err112 := mbTrans111.WriteString(arg110)
    if err112 != nil {
      Usage()
      return
    }
    factory113 := thrift.NewSimpleJSONProtocolFactory()
    jsProt114 := factory113.GetProtocol(mbTrans111)
    argvalue0 := storage.NewUpdateVertexRequest()
    err115 := argvalue0.Read(jsProt114)
    if err115 != nil {
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
    arg116 := flag.Arg(1)
    mbTrans117 := thrift.NewMemoryBufferLen(len(arg116))
    defer mbTrans117.Close()
    _, err118 := mbTrans117.WriteString(arg116)
    if err118 != nil {
      Usage()
      return
    }
    factory119 := thrift.NewSimpleJSONProtocolFactory()
    jsProt120 := factory119.GetProtocol(mbTrans117)
    argvalue0 := storage.NewUpdateEdgeRequest()
    err121 := argvalue0.Read(jsProt120)
    if err121 != nil {
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
    arg122 := flag.Arg(1)
    mbTrans123 := thrift.NewMemoryBufferLen(len(arg122))
    defer mbTrans123.Close()
    _, err124 := mbTrans123.WriteString(arg122)
    if err124 != nil {
      Usage()
      return
    }
    factory125 := thrift.NewSimpleJSONProtocolFactory()
    jsProt126 := factory125.GetProtocol(mbTrans123)
    argvalue0 := storage.NewScanVertexRequest()
    err127 := argvalue0.Read(jsProt126)
    if err127 != nil {
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
    arg128 := flag.Arg(1)
    mbTrans129 := thrift.NewMemoryBufferLen(len(arg128))
    defer mbTrans129.Close()
    _, err130 := mbTrans129.WriteString(arg128)
    if err130 != nil {
      Usage()
      return
    }
    factory131 := thrift.NewSimpleJSONProtocolFactory()
    jsProt132 := factory131.GetProtocol(mbTrans129)
    argvalue0 := storage.NewScanEdgeRequest()
    err133 := argvalue0.Read(jsProt132)
    if err133 != nil {
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
    arg134 := flag.Arg(1)
    mbTrans135 := thrift.NewMemoryBufferLen(len(arg134))
    defer mbTrans135.Close()
    _, err136 := mbTrans135.WriteString(arg134)
    if err136 != nil {
      Usage()
      return
    }
    factory137 := thrift.NewSimpleJSONProtocolFactory()
    jsProt138 := factory137.GetProtocol(mbTrans135)
    argvalue0 := storage.NewGetUUIDReq()
    err139 := argvalue0.Read(jsProt138)
    if err139 != nil {
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
    arg140 := flag.Arg(1)
    mbTrans141 := thrift.NewMemoryBufferLen(len(arg140))
    defer mbTrans141.Close()
    _, err142 := mbTrans141.WriteString(arg140)
    if err142 != nil {
      Usage()
      return
    }
    factory143 := thrift.NewSimpleJSONProtocolFactory()
    jsProt144 := factory143.GetProtocol(mbTrans141)
    argvalue0 := storage.NewLookupIndexRequest()
    err145 := argvalue0.Read(jsProt144)
    if err145 != nil {
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
    arg146 := flag.Arg(1)
    mbTrans147 := thrift.NewMemoryBufferLen(len(arg146))
    defer mbTrans147.Close()
    _, err148 := mbTrans147.WriteString(arg146)
    if err148 != nil {
      Usage()
      return
    }
    factory149 := thrift.NewSimpleJSONProtocolFactory()
    jsProt150 := factory149.GetProtocol(mbTrans147)
    argvalue0 := storage.NewLookupAndTraverseRequest()
    err151 := argvalue0.Read(jsProt150)
    if err151 != nil {
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
    arg152 := flag.Arg(1)
    mbTrans153 := thrift.NewMemoryBufferLen(len(arg152))
    defer mbTrans153.Close()
    _, err154 := mbTrans153.WriteString(arg152)
    if err154 != nil {
      Usage()
      return
    }
    factory155 := thrift.NewSimpleJSONProtocolFactory()
    jsProt156 := factory155.GetProtocol(mbTrans153)
    argvalue0 := storage.NewAddEdgesRequest()
    err157 := argvalue0.Read(jsProt156)
    if err157 != nil {
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
