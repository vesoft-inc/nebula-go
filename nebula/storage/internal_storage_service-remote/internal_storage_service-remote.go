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
  fmt.Fprintln(os.Stderr, "  ExecResponse chainAddEdges(ChainAddEdgesRequest req)")
  fmt.Fprintln(os.Stderr, "  UpdateResponse chainUpdateEdge(ChainUpdateEdgeRequest req)")
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
  client := storage.NewInternalStorageServiceClientFactory(trans, protocolFactory)
  if err := trans.Open(); err != nil {
    fmt.Fprintln(os.Stderr, "Error opening socket to ", host, ":", port, " ", err)
    os.Exit(1)
  }
  
  switch cmd {
  case "chainAddEdges":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ChainAddEdges requires 1 args")
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
    argvalue0 := storage.NewChainAddEdgesRequest()
    err307 := argvalue0.Read(jsProt306)
    if err307 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ChainAddEdges(value0))
    fmt.Print("\n")
    break
  case "chainUpdateEdge":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "ChainUpdateEdge requires 1 args")
      flag.Usage()
    }
    arg308 := flag.Arg(1)
    mbTrans309 := thrift.NewMemoryBufferLen(len(arg308))
    defer mbTrans309.Close()
    _, err310 := mbTrans309.WriteString(arg308)
    if err310 != nil {
      Usage()
      return
    }
    factory311 := thrift.NewSimpleJSONProtocolFactory()
    jsProt312 := factory311.GetProtocol(mbTrans309)
    argvalue0 := storage.NewChainUpdateEdgeRequest()
    err313 := argvalue0.Read(jsProt312)
    if err313 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.ChainUpdateEdge(value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}