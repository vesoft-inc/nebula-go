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
  fmt.Fprintln(os.Stderr, "  GetNeighborsResponse getNeighbors(GetNeighborsRequest req)")
  fmt.Fprintln(os.Stderr, "  GetPropResponse getProps(GetPropRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse addVertices(AddVerticesRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse addEdges(AddEdgesRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse deleteEdges(DeleteEdgesRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse deleteVertices(DeleteVerticesRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse deleteTags(DeleteTagsRequest req)")
  fmt.Fprintln(os.Stderr, "  UpdateResponse updateVertex(UpdateVertexRequest req)")
  fmt.Fprintln(os.Stderr, "  UpdateResponse updateEdge(UpdateEdgeRequest req)")
  fmt.Fprintln(os.Stderr, "  ScanResponse scanVertex(ScanVertexRequest req)")
  fmt.Fprintln(os.Stderr, "  ScanResponse scanEdge(ScanEdgeRequest req)")
  fmt.Fprintln(os.Stderr, "  GetUUIDResp getUUID(GetUUIDReq req)")
  fmt.Fprintln(os.Stderr, "  LookupIndexResp lookupIndex(LookupIndexRequest req)")
  fmt.Fprintln(os.Stderr, "  GetNeighborsResponse lookupAndTraverse(LookupAndTraverseRequest req)")
  fmt.Fprintln(os.Stderr, "  UpdateResponse chainUpdateEdge(UpdateEdgeRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse chainAddEdges(AddEdgesRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse chainDeleteEdges(DeleteEdgesRequest req)")
  fmt.Fprintln(os.Stderr, "  KVGetResponse get(KVGetRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse put(KVPutRequest req)")
  fmt.Fprintln(os.Stderr, "  ExecResponse remove(KVRemoveRequest req)")
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
    arg99 := flag.Arg(1)
    mbTrans100 := thrift.NewMemoryBufferLen(len(arg99))
    defer mbTrans100.Close()
    _, err101 := mbTrans100.WriteString(arg99)
    if err101 != nil {
      Usage()
      return
    }
    factory102 := thrift.NewSimpleJSONProtocolFactory()
    jsProt103 := factory102.GetProtocol(mbTrans100)
    argvalue0 := storage.NewGetNeighborsRequest()
    err104 := argvalue0.Read(jsProt103)
    if err104 != nil {
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
    arg105 := flag.Arg(1)
    mbTrans106 := thrift.NewMemoryBufferLen(len(arg105))
    defer mbTrans106.Close()
    _, err107 := mbTrans106.WriteString(arg105)
    if err107 != nil {
      Usage()
      return
    }
    factory108 := thrift.NewSimpleJSONProtocolFactory()
    jsProt109 := factory108.GetProtocol(mbTrans106)
    argvalue0 := storage.NewGetPropRequest()
    err110 := argvalue0.Read(jsProt109)
    if err110 != nil {
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
    arg111 := flag.Arg(1)
    mbTrans112 := thrift.NewMemoryBufferLen(len(arg111))
    defer mbTrans112.Close()
    _, err113 := mbTrans112.WriteString(arg111)
    if err113 != nil {
      Usage()
      return
    }
    factory114 := thrift.NewSimpleJSONProtocolFactory()
    jsProt115 := factory114.GetProtocol(mbTrans112)
    argvalue0 := storage.NewAddVerticesRequest()
    err116 := argvalue0.Read(jsProt115)
    if err116 != nil {
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
    arg117 := flag.Arg(1)
    mbTrans118 := thrift.NewMemoryBufferLen(len(arg117))
    defer mbTrans118.Close()
    _, err119 := mbTrans118.WriteString(arg117)
    if err119 != nil {
      Usage()
      return
    }
    factory120 := thrift.NewSimpleJSONProtocolFactory()
    jsProt121 := factory120.GetProtocol(mbTrans118)
    argvalue0 := storage.NewAddEdgesRequest()
    err122 := argvalue0.Read(jsProt121)
    if err122 != nil {
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
    arg123 := flag.Arg(1)
    mbTrans124 := thrift.NewMemoryBufferLen(len(arg123))
    defer mbTrans124.Close()
    _, err125 := mbTrans124.WriteString(arg123)
    if err125 != nil {
      Usage()
      return
    }
    factory126 := thrift.NewSimpleJSONProtocolFactory()
    jsProt127 := factory126.GetProtocol(mbTrans124)
    argvalue0 := storage.NewDeleteEdgesRequest()
    err128 := argvalue0.Read(jsProt127)
    if err128 != nil {
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
    arg129 := flag.Arg(1)
    mbTrans130 := thrift.NewMemoryBufferLen(len(arg129))
    defer mbTrans130.Close()
    _, err131 := mbTrans130.WriteString(arg129)
    if err131 != nil {
      Usage()
      return
    }
    factory132 := thrift.NewSimpleJSONProtocolFactory()
    jsProt133 := factory132.GetProtocol(mbTrans130)
    argvalue0 := storage.NewDeleteVerticesRequest()
    err134 := argvalue0.Read(jsProt133)
    if err134 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.DeleteVertices(value0))
    fmt.Print("\n")
    break
  case "deleteTags":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "DeleteTags requires 1 args")
      flag.Usage()
    }
    arg135 := flag.Arg(1)
    mbTrans136 := thrift.NewMemoryBufferLen(len(arg135))
    defer mbTrans136.Close()
    _, err137 := mbTrans136.WriteString(arg135)
    if err137 != nil {
      Usage()
      return
    }
    factory138 := thrift.NewSimpleJSONProtocolFactory()
    jsProt139 := factory138.GetProtocol(mbTrans136)
    argvalue0 := storage.NewDeleteTagsRequest()
    err140 := argvalue0.Read(jsProt139)
    if err140 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.DeleteTags(value0))
    fmt.Print("\n")
    break
  case "updateVertex":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "UpdateVertex requires 1 args")
      flag.Usage()
    }
    arg141 := flag.Arg(1)
    mbTrans142 := thrift.NewMemoryBufferLen(len(arg141))
    defer mbTrans142.Close()
    _, err143 := mbTrans142.WriteString(arg141)
    if err143 != nil {
      Usage()
      return
    }
    factory144 := thrift.NewSimpleJSONProtocolFactory()
    jsProt145 := factory144.GetProtocol(mbTrans142)
    argvalue0 := storage.NewUpdateVertexRequest()
    err146 := argvalue0.Read(jsProt145)
    if err146 != nil {
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
    arg147 := flag.Arg(1)
    mbTrans148 := thrift.NewMemoryBufferLen(len(arg147))
    defer mbTrans148.Close()
    _, err149 := mbTrans148.WriteString(arg147)
    if err149 != nil {
      Usage()
      return
    }
    factory150 := thrift.NewSimpleJSONProtocolFactory()
    jsProt151 := factory150.GetProtocol(mbTrans148)
    argvalue0 := storage.NewUpdateEdgeRequest()
    err152 := argvalue0.Read(jsProt151)
    if err152 != nil {
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
    arg153 := flag.Arg(1)
    mbTrans154 := thrift.NewMemoryBufferLen(len(arg153))
    defer mbTrans154.Close()
    _, err155 := mbTrans154.WriteString(arg153)
    if err155 != nil {
      Usage()
      return
    }
    factory156 := thrift.NewSimpleJSONProtocolFactory()
    jsProt157 := factory156.GetProtocol(mbTrans154)
    argvalue0 := storage.NewScanVertexRequest()
    err158 := argvalue0.Read(jsProt157)
    if err158 != nil {
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
    arg159 := flag.Arg(1)
    mbTrans160 := thrift.NewMemoryBufferLen(len(arg159))
    defer mbTrans160.Close()
    _, err161 := mbTrans160.WriteString(arg159)
    if err161 != nil {
      Usage()
      return
    }
    factory162 := thrift.NewSimpleJSONProtocolFactory()
    jsProt163 := factory162.GetProtocol(mbTrans160)
    argvalue0 := storage.NewScanEdgeRequest()
    err164 := argvalue0.Read(jsProt163)
    if err164 != nil {
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
    arg165 := flag.Arg(1)
    mbTrans166 := thrift.NewMemoryBufferLen(len(arg165))
    defer mbTrans166.Close()
    _, err167 := mbTrans166.WriteString(arg165)
    if err167 != nil {
      Usage()
      return
    }
    factory168 := thrift.NewSimpleJSONProtocolFactory()
    jsProt169 := factory168.GetProtocol(mbTrans166)
    argvalue0 := storage.NewGetUUIDReq()
    err170 := argvalue0.Read(jsProt169)
    if err170 != nil {
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
    arg171 := flag.Arg(1)
    mbTrans172 := thrift.NewMemoryBufferLen(len(arg171))
    defer mbTrans172.Close()
    _, err173 := mbTrans172.WriteString(arg171)
    if err173 != nil {
      Usage()
      return
    }
    factory174 := thrift.NewSimpleJSONProtocolFactory()
    jsProt175 := factory174.GetProtocol(mbTrans172)
    argvalue0 := storage.NewLookupIndexRequest()
    err176 := argvalue0.Read(jsProt175)
    if err176 != nil {
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
    arg177 := flag.Arg(1)
    mbTrans178 := thrift.NewMemoryBufferLen(len(arg177))
    defer mbTrans178.Close()
    _, err179 := mbTrans178.WriteString(arg177)
    if err179 != nil {
      Usage()
      return
    }
    factory180 := thrift.NewSimpleJSONProtocolFactory()
    jsProt181 := factory180.GetProtocol(mbTrans178)
    argvalue0 := storage.NewLookupAndTraverseRequest()
    err182 := argvalue0.Read(jsProt181)
    if err182 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.LookupAndTraverse(value0))
    fmt.Print("\n")
    break
  case "chainUpdateEdge":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ChainUpdateEdge requires 1 args")
      flag.Usage()
    }
    arg183 := flag.Arg(1)
    mbTrans184 := thrift.NewMemoryBufferLen(len(arg183))
    defer mbTrans184.Close()
    _, err185 := mbTrans184.WriteString(arg183)
    if err185 != nil {
      Usage()
      return
    }
    factory186 := thrift.NewSimpleJSONProtocolFactory()
    jsProt187 := factory186.GetProtocol(mbTrans184)
    argvalue0 := storage.NewUpdateEdgeRequest()
    err188 := argvalue0.Read(jsProt187)
    if err188 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ChainUpdateEdge(value0))
    fmt.Print("\n")
    break
  case "chainAddEdges":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ChainAddEdges requires 1 args")
      flag.Usage()
    }
    arg189 := flag.Arg(1)
    mbTrans190 := thrift.NewMemoryBufferLen(len(arg189))
    defer mbTrans190.Close()
    _, err191 := mbTrans190.WriteString(arg189)
    if err191 != nil {
      Usage()
      return
    }
    factory192 := thrift.NewSimpleJSONProtocolFactory()
    jsProt193 := factory192.GetProtocol(mbTrans190)
    argvalue0 := storage.NewAddEdgesRequest()
    err194 := argvalue0.Read(jsProt193)
    if err194 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ChainAddEdges(value0))
    fmt.Print("\n")
    break
  case "chainDeleteEdges":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ChainDeleteEdges requires 1 args")
      flag.Usage()
    }
    arg195 := flag.Arg(1)
    mbTrans196 := thrift.NewMemoryBufferLen(len(arg195))
    defer mbTrans196.Close()
    _, err197 := mbTrans196.WriteString(arg195)
    if err197 != nil {
      Usage()
      return
    }
    factory198 := thrift.NewSimpleJSONProtocolFactory()
    jsProt199 := factory198.GetProtocol(mbTrans196)
    argvalue0 := storage.NewDeleteEdgesRequest()
    err200 := argvalue0.Read(jsProt199)
    if err200 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ChainDeleteEdges(value0))
    fmt.Print("\n")
    break
  case "get":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "Get requires 1 args")
      flag.Usage()
    }
    arg201 := flag.Arg(1)
    mbTrans202 := thrift.NewMemoryBufferLen(len(arg201))
    defer mbTrans202.Close()
    _, err203 := mbTrans202.WriteString(arg201)
    if err203 != nil {
      Usage()
      return
    }
    factory204 := thrift.NewSimpleJSONProtocolFactory()
    jsProt205 := factory204.GetProtocol(mbTrans202)
    argvalue0 := storage.NewKVGetRequest()
    err206 := argvalue0.Read(jsProt205)
    if err206 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.Get(value0))
    fmt.Print("\n")
    break
  case "put":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "Put requires 1 args")
      flag.Usage()
    }
    arg207 := flag.Arg(1)
    mbTrans208 := thrift.NewMemoryBufferLen(len(arg207))
    defer mbTrans208.Close()
    _, err209 := mbTrans208.WriteString(arg207)
    if err209 != nil {
      Usage()
      return
    }
    factory210 := thrift.NewSimpleJSONProtocolFactory()
    jsProt211 := factory210.GetProtocol(mbTrans208)
    argvalue0 := storage.NewKVPutRequest()
    err212 := argvalue0.Read(jsProt211)
    if err212 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.Put(value0))
    fmt.Print("\n")
    break
  case "remove":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "Remove requires 1 args")
      flag.Usage()
    }
    arg213 := flag.Arg(1)
    mbTrans214 := thrift.NewMemoryBufferLen(len(arg213))
    defer mbTrans214.Close()
    _, err215 := mbTrans214.WriteString(arg213)
    if err215 != nil {
      Usage()
      return
    }
    factory216 := thrift.NewSimpleJSONProtocolFactory()
    jsProt217 := factory216.GetProtocol(mbTrans214)
    argvalue0 := storage.NewKVRemoveRequest()
    err218 := argvalue0.Read(jsProt217)
    if err218 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.Remove(value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
