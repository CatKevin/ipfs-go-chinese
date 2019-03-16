
//<developer>
//    <name>linapex 曹一峰</name>
//    <email>linapex@163.com</email>
//    <wx>superexc</wx>
//    <qqgroup>128148617</qqgroup>
//    <url>https://jsq.ink</url>
//    <role>pku engineer</role>
//    <date>2019-03-16 19:56:41</date>
//</624460172697276416>

/*
包correpo为本地提供pinning和垃圾收集
IPFS块服务。

IPFS节点将保留任何对象的本地副本
本地添加或请求。不是所有这些东西都值得
但是要永久保存，以便节点管理员可以固定对象
他们想保留和取消固定他们不关心的对象。

垃圾收集扫描循环访问本地块存储
删除未固定的对象，从而释放新的存储空间
物体。
**/

package corerepo

import (
	"context"
	"fmt"
	"github.com/ipfs/go-ipfs/pin"

	"github.com/ipfs/go-ipfs/core/coreapi/interface"

	"gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
)

func Pin(pinning pin.Pinner, api iface.CoreAPI, ctx context.Context, paths []string, recursive bool) ([]cid.Cid, error) {
	out := make([]cid.Cid, len(paths))

	for i, fpath := range paths {
		p, err := iface.ParsePath(fpath)
		if err != nil {
			return nil, err
		}

		dagnode, err := api.ResolveNode(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("pin: %s", err)
		}
		err = pinning.Pin(ctx, dagnode, recursive)
		if err != nil {
			return nil, fmt.Errorf("pin: %s", err)
		}
		out[i] = dagnode.Cid()
	}

	err := pinning.Flush()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func Unpin(pinning pin.Pinner, api iface.CoreAPI, ctx context.Context, paths []string, recursive bool) ([]cid.Cid, error) {
	unpinned := make([]cid.Cid, len(paths))

	for i, p := range paths {
		p, err := iface.ParsePath(p)
		if err != nil {
			return nil, err
		}

		k, err := api.ResolvePath(ctx, p)
		if err != nil {
			return nil, err
		}

		err = pinning.Unpin(ctx, k.Cid(), recursive)
		if err != nil {
			return nil, err
		}
		unpinned[i] = k.Cid()
	}

	err := pinning.Flush()
	if err != nil {
		return nil, err
	}
	return unpinned, nil
}

