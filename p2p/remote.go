
//<developer>
//    <name>linapex 曹一峰</name>
//    <email>linapex@163.com</email>
//    <wx>superexc</wx>
//    <qqgroup>128148617</qqgroup>
//    <url>https://jsq.ink</url>
//    <role>pku engineer</role>
//    <date>2019-03-16 19:56:43</date>
//</624460181278822400>

package p2p

import (
	"context"
	"fmt"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	net "gx/ipfs/QmNgLg1NTw37iWbYPKcyK85YJ9Whs1MkPtJwhfqbNYAyKg/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	manet "gx/ipfs/QmZcLBXKaFe8ND5YHPkJRAwmhJGrVsi1JqDZNyJ4nRK5Mj/go-multiaddr-net"
)

var maPrefix = "/" + ma.ProtocolWithCode(ma.P_IPFS).Name + "/"

//远程侦听器接受libp2p流并将其代理到manet主机
type remoteListener struct {
	p2p *P2P

//应用程序协议标识符。
	proto protocol.ID

//代理传入连接的地址
	addr ma.Multiaddr

//reportremote if set to true makes the handler send'<base58 remote peerid>\n'
//在转发任何数据之前指向目标
	reportRemote bool
}

//ForwardRemote创建新的P2P侦听器
func (p2p *P2P) ForwardRemote(ctx context.Context, proto protocol.ID, addr ma.Multiaddr, reportRemote bool) (Listener, error) {
	listener := &remoteListener{
		p2p: p2p,

		proto: proto,
		addr:  addr,

		reportRemote: reportRemote,
	}

	if err := p2p.ListenersP2P.Register(listener); err != nil {
		return nil, err
	}

	return listener, nil
}

func (l *remoteListener) handleStream(remote net.Stream) {
	local, err := manet.Dial(l.addr)
	if err != nil {
		remote.Reset()
		return
	}

	peer := remote.Conn().RemotePeer()

	if l.reportRemote {
		if _, err := fmt.Fprintf(local, "%s\n", peer.Pretty()); err != nil {
			remote.Reset()
			return
		}
	}

	peerMa, err := ma.NewMultiaddr(maPrefix + peer.Pretty())
	if err != nil {
		remote.Reset()
		return
	}

	stream := &Stream{
		Protocol: l.proto,

		OriginAddr: peerMa,
		TargetAddr: l.addr,
		peer:       peer,

		Local:  local,
		Remote: remote,

		Registry: l.p2p.Streams,
	}

	l.p2p.Streams.Register(stream)
}

func (l *remoteListener) Protocol() protocol.ID {
	return l.proto
}

func (l *remoteListener) ListenAddress() ma.Multiaddr {
	addr, err := ma.NewMultiaddr(maPrefix + l.p2p.identity.Pretty())
	if err != nil {
		panic(err)
	}
	return addr
}

func (l *remoteListener) TargetAddress() ma.Multiaddr {
	return l.addr
}

func (l *remoteListener) close() {}

func (l *remoteListener) key() string {
	return string(l.proto)
}

