
//<developer>
//    <name>linapex 曹一峰</name>
//    <email>linapex@163.com</email>
//    <wx>superexc</wx>
//    <qqgroup>128148617</qqgroup>
//    <url>https://jsq.ink</url>
//    <role>pku engineer</role>
//    <date>2019-03-16 19:56:39</date>
//</624460161720782848>

//+建设！窗户，！诺福斯

package commands

import (
	"fmt"
	"io"

	cmdenv "github.com/ipfs/go-ipfs/core/commands/cmdenv"
	nodeMount "github.com/ipfs/go-ipfs/fuse/node"

	cmds "gx/ipfs/QmWGm4AbZEbnmdgVTza52MSNpEmBdFVqzmAysRbjrRyGbH/go-ipfs-cmds"
	config "gx/ipfs/QmcRKBUqc2p3L1ZraoJjbXfs9E6xzvEuyK9iypb5RGwfsr/go-ipfs-config"
	cmdkit "gx/ipfs/Qmde5VP1qUkyQXKCfmEUA7bP64V2HAptbJ7phuPp7jXWwg/go-ipfs-cmdkit"
)

const (
	mountIPFSPathOptionName = "ipfs-path"
	mountIPNSPathOptionName = "ipns-path"
)

var MountCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Mounts IPFS to the filesystem (read-only).",
		ShortDescription: `
Mount IPFS at a read-only mountpoint on the OS (default: /ipfs and /ipns).
All IPFS objects will be accessible under that directory. Note that the
root will not be listable, as it is virtual. Access known paths directly.

You may have to create /ipfs and /ipns before using 'ipfs mount':

> sudo mkdir /ipfs /ipns
> sudo chown $(whoami) /ipfs /ipns
> ipfs daemon &
> ipfs mount
`,
		LongDescription: `
Mount IPFS at a read-only mountpoint on the OS. The default, /ipfs and /ipns,
are set in the configuration file, but can be overriden by the options.
All IPFS objects will be accessible under this directory. Note that the
root will not be listable, as it is virtual. Access known paths directly.

You may have to create /ipfs and /ipns before using 'ipfs mount':

> sudo mkdir /ipfs /ipns
> sudo chown $(whoami) /ipfs /ipns
> ipfs daemon &
> ipfs mount

Example:

# setup
> mkdir foo
> echo "baz" > foo/bar
> ipfs add -r foo
added QmWLdkp93sNxGRjnFHPaYg8tCQ35NBY3XPn6KiETd3Z4WR foo/bar
added QmSh5e7S6fdcu75LAbXNZAFY2nGyZUJXyLCJDvn2zRkWyC foo
> ipfs ls QmSh5e7S6fdcu75LAbXNZAFY2nGyZUJXyLCJDvn2zRkWyC
QmWLdkp93sNxGRjnFHPaYg8tCQ35NBY3XPn6KiETd3Z4WR 12 bar
> ipfs cat QmWLdkp93sNxGRjnFHPaYg8tCQ35NBY3XPn6KiETd3Z4WR
baz

# mount
> ipfs daemon &
> ipfs mount
IPFS mounted at: /ipfs
IPNS mounted at: /ipns
> cd /ipfs/QmSh5e7S6fdcu75LAbXNZAFY2nGyZUJXyLCJDvn2zRkWyC
> ls
bar
> cat bar
baz
> cat /ipfs/QmSh5e7S6fdcu75LAbXNZAFY2nGyZUJXyLCJDvn2zRkWyC/bar
baz
> cat /ipfs/QmWLdkp93sNxGRjnFHPaYg8tCQ35NBY3XPn6KiETd3Z4WR
baz
`,
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption(mountIPFSPathOptionName, "f", "The path where IPFS should be mounted."),
		cmdkit.StringOption(mountIPNSPathOptionName, "n", "The path where IPNS should be mounted."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cfg, err := cmdenv.GetConfig(env)
		if err != nil {
			return err
		}

		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

//如果我们没有在联机模式下运行节点，则会出错
		if nd.LocalMode() {
			return ErrNotOnline
		}

		fsdir, found := req.Options[mountIPFSPathOptionName].(string)
		if !found {
fsdir = cfg.Mounts.IPFS //使用默认值
		}

//获取默认安装点
		nsdir, found := req.Options[mountIPNSPathOptionName].(string)
		if !found {
nsdir = cfg.Mounts.IPNS //注意：一定不要重新申报！
		}

		err = nodeMount.Mount(nd, fsdir, nsdir)
		if err != nil {
			return err
		}

		var output config.Mounts
		output.IPFS = fsdir
		output.IPNS = nsdir
		return cmds.EmitOnce(res, &output)
	},
	Type: config.Mounts{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, mounts *config.Mounts) error {
			fmt.Fprintf(w, "IPFS mounted at: %s\n", mounts.IPFS)
			fmt.Fprintf(w, "IPNS mounted at: %s\n", mounts.IPNS)

			return nil
		}),
	},
}

