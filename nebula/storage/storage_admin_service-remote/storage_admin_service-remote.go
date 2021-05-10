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

	"../../github.com/vesoft-inc/nebula-go/v2/nebula/storage"

	thrift "github.com/facebook/fbthrift/thrift/lib/go/thrift"
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
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "TransLeader requires 1 args")
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
		argvalue0 := storage.NewTransLeaderReq()
		err164 := argvalue0.Read(jsProt163)
		if err164 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.TransLeader(value0))
		fmt.Print("\n")
		break
	case "addPart":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "AddPart requires 1 args")
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
		argvalue0 := storage.NewAddPartReq()
		err170 := argvalue0.Read(jsProt169)
		if err170 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.AddPart(value0))
		fmt.Print("\n")
		break
	case "addLearner":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "AddLearner requires 1 args")
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
		argvalue0 := storage.NewAddLearnerReq()
		err176 := argvalue0.Read(jsProt175)
		if err176 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.AddLearner(value0))
		fmt.Print("\n")
		break
	case "removePart":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "RemovePart requires 1 args")
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
		argvalue0 := storage.NewRemovePartReq()
		err182 := argvalue0.Read(jsProt181)
		if err182 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.RemovePart(value0))
		fmt.Print("\n")
		break
	case "memberChange":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "MemberChange requires 1 args")
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
		argvalue0 := storage.NewMemberChangeReq()
		err188 := argvalue0.Read(jsProt187)
		if err188 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.MemberChange(value0))
		fmt.Print("\n")
		break
	case "waitingForCatchUpData":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "WaitingForCatchUpData requires 1 args")
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
		argvalue0 := storage.NewCatchUpDataReq()
		err194 := argvalue0.Read(jsProt193)
		if err194 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.WaitingForCatchUpData(value0))
		fmt.Print("\n")
		break
	case "createCheckpoint":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "CreateCheckpoint requires 1 args")
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
		argvalue0 := storage.NewCreateCPRequest()
		err200 := argvalue0.Read(jsProt199)
		if err200 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.CreateCheckpoint(value0))
		fmt.Print("\n")
		break
	case "dropCheckpoint":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "DropCheckpoint requires 1 args")
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
		argvalue0 := storage.NewDropCPRequest()
		err206 := argvalue0.Read(jsProt205)
		if err206 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.DropCheckpoint(value0))
		fmt.Print("\n")
		break
	case "blockingWrites":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "BlockingWrites requires 1 args")
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
		argvalue0 := storage.NewBlockingSignRequest()
		err212 := argvalue0.Read(jsProt211)
		if err212 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.BlockingWrites(value0))
		fmt.Print("\n")
		break
	case "rebuildTagIndex":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "RebuildTagIndex requires 1 args")
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
		argvalue0 := storage.NewRebuildIndexRequest()
		err218 := argvalue0.Read(jsProt217)
		if err218 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.RebuildTagIndex(value0))
		fmt.Print("\n")
		break
	case "rebuildEdgeIndex":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "RebuildEdgeIndex requires 1 args")
			flag.Usage()
		}
		arg219 := flag.Arg(1)
		mbTrans220 := thrift.NewMemoryBufferLen(len(arg219))
		defer mbTrans220.Close()
		_, err221 := mbTrans220.WriteString(arg219)
		if err221 != nil {
			Usage()
			return
		}
		factory222 := thrift.NewSimpleJSONProtocolFactory()
		jsProt223 := factory222.GetProtocol(mbTrans220)
		argvalue0 := storage.NewRebuildIndexRequest()
		err224 := argvalue0.Read(jsProt223)
		if err224 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.RebuildEdgeIndex(value0))
		fmt.Print("\n")
		break
	case "getLeaderParts":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "GetLeaderParts requires 1 args")
			flag.Usage()
		}
		arg225 := flag.Arg(1)
		mbTrans226 := thrift.NewMemoryBufferLen(len(arg225))
		defer mbTrans226.Close()
		_, err227 := mbTrans226.WriteString(arg225)
		if err227 != nil {
			Usage()
			return
		}
		factory228 := thrift.NewSimpleJSONProtocolFactory()
		jsProt229 := factory228.GetProtocol(mbTrans226)
		argvalue0 := storage.NewGetLeaderReq()
		err230 := argvalue0.Read(jsProt229)
		if err230 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.GetLeaderParts(value0))
		fmt.Print("\n")
		break
	case "checkPeers":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "CheckPeers requires 1 args")
			flag.Usage()
		}
		arg231 := flag.Arg(1)
		mbTrans232 := thrift.NewMemoryBufferLen(len(arg231))
		defer mbTrans232.Close()
		_, err233 := mbTrans232.WriteString(arg231)
		if err233 != nil {
			Usage()
			return
		}
		factory234 := thrift.NewSimpleJSONProtocolFactory()
		jsProt235 := factory234.GetProtocol(mbTrans232)
		argvalue0 := storage.NewCheckPeersReq()
		err236 := argvalue0.Read(jsProt235)
		if err236 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.CheckPeers(value0))
		fmt.Print("\n")
		break
	case "addAdminTask":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "AddAdminTask requires 1 args")
			flag.Usage()
		}
		arg237 := flag.Arg(1)
		mbTrans238 := thrift.NewMemoryBufferLen(len(arg237))
		defer mbTrans238.Close()
		_, err239 := mbTrans238.WriteString(arg237)
		if err239 != nil {
			Usage()
			return
		}
		factory240 := thrift.NewSimpleJSONProtocolFactory()
		jsProt241 := factory240.GetProtocol(mbTrans238)
		argvalue0 := storage.NewAddAdminTaskRequest()
		err242 := argvalue0.Read(jsProt241)
		if err242 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.AddAdminTask(value0))
		fmt.Print("\n")
		break
	case "stopAdminTask":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "StopAdminTask requires 1 args")
			flag.Usage()
		}
		arg243 := flag.Arg(1)
		mbTrans244 := thrift.NewMemoryBufferLen(len(arg243))
		defer mbTrans244.Close()
		_, err245 := mbTrans244.WriteString(arg243)
		if err245 != nil {
			Usage()
			return
		}
		factory246 := thrift.NewSimpleJSONProtocolFactory()
		jsProt247 := factory246.GetProtocol(mbTrans244)
		argvalue0 := storage.NewStopAdminTaskRequest()
		err248 := argvalue0.Read(jsProt247)
		if err248 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.StopAdminTask(value0))
		fmt.Print("\n")
		break
	case "":
		Usage()
		break
	default:
		fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
	}
}
