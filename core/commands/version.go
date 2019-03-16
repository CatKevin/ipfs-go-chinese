
//<developer>
//    <name>linapex 曹一峰</name>
//    <email>linapex@163.com</email>
//    <wx>superexc</wx>
//    <qqgroup>128148617</qqgroup>
//    <url>https://jsq.ink</url>
//    <role>pku engineer</role>
//    <date>2019-03-16 19:56:39</date>
//</624460164518383616>

package commands

import (
	"fmt"
	"io"
	"runtime"

	version "github.com/ipfs/go-ipfs"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	cmds "gx/ipfs/QmWGm4AbZEbnmdgVTza52MSNpEmBdFVqzmAysRbjrRyGbH/go-ipfs-cmds"
	"gx/ipfs/Qmde5VP1qUkyQXKCfmEUA7bP64V2HAptbJ7phuPp7jXWwg/go-ipfs-cmdkit"
)

type VersionOutput struct {
	Version string
	Commit  string
	Repo    string
	System  string
	Golang  string
}

const (
	versionNumberOptionName = "number"
	versionCommitOptionName = "commit"
	versionRepoOptionName   = "repo"
	versionAllOptionName    = "all"
)

var VersionCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline:          "Show ipfs version information.",
		ShortDescription: "Returns the current version of ipfs and exits.",
	},

	Options: []cmdkit.Option{
		cmdkit.BoolOption(versionNumberOptionName, "n", "Only show the version number."),
		cmdkit.BoolOption(versionCommitOptionName, "Show the commit hash."),
		cmdkit.BoolOption(versionRepoOptionName, "Show repo version."),
		cmdkit.BoolOption(versionAllOptionName, "Show all version information"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		return cmds.EmitOnce(res, &VersionOutput{
			Version: version.CurrentVersionNumber,
			Commit:  version.CurrentCommit,
			Repo:    fmt.Sprint(fsrepo.RepoVersion),
System:  runtime.GOARCH + "/" + runtime.GOOS, //TODO:此处为精确版本
			Golang:  runtime.Version(),
		})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, version *VersionOutput) error {
			commit, _ := req.Options[versionCommitOptionName].(bool)
			commitTxt := ""
			if commit {
				commitTxt = "-" + version.Commit
			}

			all, _ := req.Options[versionAllOptionName].(bool)
			if all {
				out := fmt.Sprintf("go-ipfs version: %s-%s\n"+
					"Repo version: %s\nSystem version: %s\nGolang version: %s\n",
					version.Version, version.Commit, version.Repo, version.System, version.Golang)
				fmt.Fprint(w, out)
				return nil
			}

			repo, _ := req.Options[versionRepoOptionName].(bool)
			if repo {
				fmt.Fprintln(w, version.Repo)
				return nil
			}

			number, _ := req.Options[versionNumberOptionName].(bool)
			if number {
				fmt.Fprintln(w, version.Version+commitTxt)
				return nil
			}

			fmt.Fprint(w, fmt.Sprintf("ipfs version %s%s\n", version.Version, commitTxt))
			return nil
		}),
	},
	Type: VersionOutput{},
}

