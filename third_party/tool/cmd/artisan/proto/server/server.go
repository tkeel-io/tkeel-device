package server

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"

	"github.com/emicklei/proto"
	"github.com/spf13/cobra"
)

// CmdServer the service command.
var CmdServer = &cobra.Command{
	Use:   "server",
	Short: "Generate the proto Server implementations",
	Long:  "Generate the proto Server implementations. Example: tkeel proto server api/xxx.proto -target-dir=pkg/service",
	Run:   run,
}
var targetDir string

func init() {
	CmdServer.Flags().StringVarP(&targetDir, "target-dir", "t", "pkg/service", "generate target directory")
}

func run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify the proto file. Example: tkeel proto server api/xxx.proto")
		return
	}
	reader, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	var (
		pkg string
		res []*Service
	)
	proto.Walk(definition,
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" {
				pkg = strings.Split(o.Constant.Source, ";")[0]
			}
		}),
		proto.WithService(func(s *proto.Service) {
			cs := &Service{
				Package: pkg,
				Service: s.Name,
			}
			for _, e := range s.Elements {
				r, ok := e.(*proto.RPC)
				if ok {
					cs.Methods = append(cs.Methods, &Method{
						Service: s.Name, Name: r.Name, Request: r.RequestType,
						Reply: r.ReturnsType, Type: getMethodType(r.StreamsRequest, r.StreamsReturns),
					})
				}
			}
			res = append(res, cs)
		}),
	)

	fmt.Print("💻 Add the following code to cmd/<project>.go  👇:\n\n")
	for _, s := range res {
		fmt.Println(color.WhiteString("import("))
		fmt.Println(color.WhiteString("%s_v1 \"%s\"", s.Service, s.Package))
		fmt.Println(color.WhiteString(")"))
		fmt.Println()
		fmt.Println(color.WhiteString("%sSrv := service.New%sService()", s.Service, s.Service))
		fmt.Println(color.WhiteString("%s_v1.Register%sHTTPServer(httpSrv.Container, %sSrv)", s.Service, s.Service, s.Service))
		fmt.Println(color.WhiteString("%s_v1.Register%sServer(grpcSrv.GetServe(), %sSrv)", s.Service, s.Service, s.Service))
	}
}

func getMethodType(streamsRequest, streamsReturns bool) MethodType {
	if !streamsRequest && !streamsReturns {
		return unaryType
	} else if streamsRequest && streamsReturns {
		return twoWayStreamsType
	} else if streamsRequest {
		return requestStreamsType
	} else if streamsReturns {
		return returnsStreamsType
	}
	return unaryType
}
