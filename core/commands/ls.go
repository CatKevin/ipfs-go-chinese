
//<developer>
//    <name>linapex 曹一峰</name>
//    <email>linapex@163.com</email>
//    <wx>superexc</wx>
//    <qqgroup>128148617</qqgroup>
//    <url>https://jsq.ink</url>
//    <role>pku engineer</role>
//    <date>2019-03-16 19:56:39</date>
//</624460161565593600>

package commands

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	cmdenv "github.com/ipfs/go-ipfs/core/commands/cmdenv"
	iface "github.com/ipfs/go-ipfs/core/coreapi/interface"

	unixfs "gx/ipfs/QmQXze9tG878pa4Euya4rrDpyTNX3kQe4dhCaBzBozGgpe/go-unixfs"
	uio "gx/ipfs/QmQXze9tG878pa4Euya4rrDpyTNX3kQe4dhCaBzBozGgpe/go-unixfs/io"
	unixfspb "gx/ipfs/QmQXze9tG878pa4Euya4rrDpyTNX3kQe4dhCaBzBozGgpe/go-unixfs/pb"
	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	merkledag "gx/ipfs/QmTQdH4848iTVCJmKXYyRiK72HufWTLYQQ8iN3JaQ8K1Hq/go-merkledag"
	cmds "gx/ipfs/QmWGm4AbZEbnmdgVTza52MSNpEmBdFVqzmAysRbjrRyGbH/go-ipfs-cmds"
	blockservice "gx/ipfs/QmYPZzd9VqmJDwxUnThfeSbV1Y5o53aVPDijTB7j7rS9Ep/go-blockservice"
	offline "gx/ipfs/QmYZwey1thDTynSrvd6qQkX24UpTka6TFhQ2v569UpoqxD/go-ipfs-exchange-offline"
	ipld "gx/ipfs/QmcKKBwfz6FyQdHR2jsXrrF6XeSBXYL86anmWNewpFpoF5/go-ipld-format"
	"gx/ipfs/Qmde5VP1qUkyQXKCfmEUA7bP64V2HAptbJ7phuPp7jXWwg/go-ipfs-cmdkit"
)

//ls link包含ls输出中单个ipld链接的可打印数据
type LsLink struct {
	Name, Hash string
	Size       uint64
	Type       unixfspb.Data_DataType
}

//lsobject是lsoutput的元素
//它可以表示目录的全部或部分
type LsObject struct {
	Hash  string
	Links []LsLink
}

//lsoutput是一组可打印的目录数据，
//它可以是完整的，也可以是部分的
type LsOutput struct {
	Objects []LsObject
}

const (
	lsHeadersOptionNameTime = "headers"
	lsResolveTypeOptionName = "resolve-type"
	lsSizeOptionName        = "size"
	lsStreamOptionName      = "stream"
)

var LsCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "List directory contents for Unix filesystem objects.",
		ShortDescription: `
Displays the contents of an IPFS or IPNS object(s) at the given path, with
the following format:

  <link base58 hash> <link size in bytes> <link name>

The JSON output contains type information.
`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("ipfs-path", true, true, "The path to the IPFS object(s) to list links from.").EnableStdin(),
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption(lsHeadersOptionNameTime, "v", "Print table headers (Hash, Size, Name)."),
		cmdkit.BoolOption(lsResolveTypeOptionName, "Resolve linked objects to find out their types.").WithDefault(true),
		cmdkit.BoolOption(lsSizeOptionName, "Resolve linked objects to find out their file size.").WithDefault(true),
		cmdkit.BoolOption(lsStreamOptionName, "s", "Enable exprimental streaming of directory entries as they are traversed."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}

		resolveType, _ := req.Options[lsResolveTypeOptionName].(bool)
		resolveSize, _ := req.Options[lsSizeOptionName].(bool)
		dserv := nd.DAG
		if !resolveType && !resolveSize {
			offlineexch := offline.Exchange(nd.Blockstore)
			bserv := blockservice.New(nd.Blockstore, offlineexch)
			dserv = merkledag.NewDAGService(bserv)
		}

		err = req.ParseBodyArgs()
		if err != nil {
			return err
		}

		paths := req.Arguments

		var dagnodes []ipld.Node
		for _, fpath := range paths {
			p, err := iface.ParsePath(fpath)
			if err != nil {
				return err
			}
			dagnode, err := api.ResolveNode(req.Context, p)
			if err != nil {
				return err
			}
			dagnodes = append(dagnodes, dagnode)
		}
		ng := merkledag.NewSession(req.Context, nd.DAG)
		ro := merkledag.NewReadOnlyDagService(ng)

		stream, _ := req.Options[lsStreamOptionName].(bool)

		if !stream {
			output := make([]LsObject, len(req.Arguments))

			for i, dagnode := range dagnodes {
				dir, err := uio.NewDirectoryFromNode(ro, dagnode)
				if err != nil && err != uio.ErrNotADir {
					return fmt.Errorf("the data in %s (at %q) is not a UnixFS directory: %s", dagnode.Cid(), paths[i], err)
				}

				var links []*ipld.Link
				if dir == nil {
					links = dagnode.Links()
				} else {
					links, err = dir.Links(req.Context)
					if err != nil {
						return err
					}
				}
				outputLinks := make([]LsLink, len(links))
				for j, link := range links {
					lsLink, err := makeLsLink(req, dserv, resolveType, resolveSize, link)
					if err != nil {
						return err
					}
					outputLinks[j] = *lsLink
				}
				output[i] = LsObject{
					Hash:  paths[i],
					Links: outputLinks,
				}
			}

			return cmds.EmitOnce(res, &LsOutput{output})
		}

		for i, dagnode := range dagnodes {
			dir, err := uio.NewDirectoryFromNode(ro, dagnode)
			if err != nil && err != uio.ErrNotADir {
				return fmt.Errorf("the data in %s (at %q) is not a UnixFS directory: %s", dagnode.Cid(), paths[i], err)
			}

			var linkResults <-chan unixfs.LinkResult
			if dir == nil {
				linkResults = makeDagNodeLinkResults(req, dagnode)
			} else {
				linkResults = dir.EnumLinksAsync(req.Context)
			}

			for linkResult := range linkResults {

				if linkResult.Err != nil {
					return linkResult.Err
				}
				link := linkResult.Link
				lsLink, err := makeLsLink(req, dserv, resolveType, resolveSize, link)
				if err != nil {
					return err
				}
				output := []LsObject{{
					Hash:  paths[i],
					Links: []LsLink{*lsLink},
				}}
				if err = res.Emit(&LsOutput{output}); err != nil {
					return err
				}
			}
		}
		return nil
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			req := res.Request()
			lastObjectHash := ""

			for {
				v, err := res.Next()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
				out := v.(*LsOutput)
				lastObjectHash = tabularOutput(req, os.Stdout, out, lastObjectHash, false)
			}
		},
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *LsOutput) error {
//使用文本编码器在HTTP上进行流式处理时，无法呈现中断
//因为我们不知道最后一个的哈希值
//目录编码器
			ignoreBreaks, _ := req.Options[lsStreamOptionName].(bool)
			tabularOutput(req, w, out, "", ignoreBreaks)
			return nil
		}),
	},
	Type: LsOutput{},
}

func makeDagNodeLinkResults(req *cmds.Request, dagnode ipld.Node) <-chan unixfs.LinkResult {
	links := dagnode.Links()
	linkResults := make(chan unixfs.LinkResult, len(links))
	defer close(linkResults)
	for _, l := range links {
		linkResults <- unixfs.LinkResult{
			Link: l,
			Err:  nil,
		}
	}
	return linkResults
}

func makeLsLink(req *cmds.Request, dserv ipld.DAGService, resolveType bool, resolveSize bool, link *ipld.Link) (*LsLink, error) {
	t := unixfspb.Data_DataType(-1)
	var size uint64

	switch link.Cid.Type() {
	case cid.Raw:
//不需要检查生叶子
		t = unixfs.TFile
		size = link.Size
	case cid.DagProtobuf:
		linkNode, err := link.GetNode(req.Context, dserv)
		if err == ipld.ErrNotFound && !resolveType && !resolveSize {
//不是错误
			linkNode = nil
		} else if err != nil {
			return nil, err
		}

		if pn, ok := linkNode.(*merkledag.ProtoNode); ok {
			d, err := unixfs.FSNodeFromBytes(pn.Data())
			if err != nil {
				return nil, err
			}
			if resolveType {
				t = d.Type()
			}
			if d.Type() == unixfs.TFile && resolveSize {
				size = d.FileSize()
			}
		}
	}
	return &LsLink{
		Name: link.Name,
		Hash: link.Cid.String(),
		Size: size,
		Type: t,
	}, nil
}

func tabularOutput(req *cmds.Request, w io.Writer, out *LsOutput, lastObjectHash string, ignoreBreaks bool) string {
	headers, _ := req.Options[lsHeadersOptionNameTime].(bool)
	stream, _ := req.Options[lsStreamOptionName].(bool)
	size, _ := req.Options[lsSizeOptionName].(bool)
//在流模式下，我们无法自动对齐选项卡
//所以我们最好猜猜
	var minTabWidth int
	if stream {
		minTabWidth = 10
	} else {
		minTabWidth = 1
	}

	multipleFolders := len(req.Arguments) > 1

	tw := tabwriter.NewWriter(w, minTabWidth, 2, 1, ' ', 0)

	for _, object := range out.Objects {

		if !ignoreBreaks && object.Hash != lastObjectHash {
			if multipleFolders {
				if lastObjectHash != "" {
					fmt.Fprintln(tw)
				}
				fmt.Fprintf(tw, "%s:\n", object.Hash)
			}
			if headers {
				s := "Hash\tName"
				if size {
					s = "Hash\tSize\tName"
				}
				fmt.Fprintln(tw, s)
			}
			lastObjectHash = object.Hash
		}

		for _, link := range object.Links {
			s := "%[1]s\t%[3]s\n"

			switch {
			case link.Type == unixfs.TDirectory && size:
				s = "%[1]s\t-\t%[3]s/\n"
			case link.Type == unixfs.TDirectory && !size:
				s = "%[1]s\t%[3]s/\n"
			case size:
				s = "%s\t%v\t%s\n"
			}

			fmt.Fprintf(tw, s, link.Hash, link.Size, link.Name)
		}
	}
	tw.Flush()
	return lastObjectHash
}

