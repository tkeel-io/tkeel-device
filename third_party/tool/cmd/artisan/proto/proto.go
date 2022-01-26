package proto

import (
	"github.com/tkeel-io/tkeel-interface/tool/cmd/artisan/proto/add"
	"github.com/tkeel-io/tkeel-interface/tool/cmd/artisan/proto/client"
	"github.com/tkeel-io/tkeel-interface/tool/cmd/artisan/proto/server"
	"github.com/tkeel-io/tkeel-interface/tool/cmd/artisan/proto/service"

	"github.com/spf13/cobra"
)

// CmdProto represents the proto command.
var CmdProto = &cobra.Command{
	Use:   "proto",
	Short: "Generate the proto files",
	Long:  "Generate the proto files.",
	Run:   run,
}

func init() {
	CmdProto.AddCommand(add.CmdAdd)
	CmdProto.AddCommand(client.CmdClient)
	CmdProto.AddCommand(service.CmdService)
	CmdProto.AddCommand(server.CmdServer)
}

func run(cmd *cobra.Command, args []string) {
	// Prompt help information If there is no parameter
	if len(args) == 0 {
		cmd.Help()
		return
	}
}
