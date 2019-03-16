
//<developer>
//    <name>linapex 曹一峰</name>
//    <email>linapex@163.com</email>
//    <wx>superexc</wx>
//    <qqgroup>128148617</qqgroup>
//    <url>https://jsq.ink</url>
//    <role>pku engineer</role>
//    <date>2019-03-16 19:56:41</date>
//</624460169262141440>

package coreapi

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sort"

	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	caopts "github.com/ipfs/go-ipfs/core/coreapi/interface/options"

	ipfspath "gx/ipfs/QmNYPETsdAu2uQ1k9q9S1jYEGURaLHV6cbYRSVFVRftpF8/go-path"
	crypto "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
)

type KeyAPI CoreAPI

type key struct {
	name   string
	peerID peer.ID
}

//name返回密钥名
func (k *key) Name() string {
	return k.name
}

//path返回键的路径。
func (k *key) Path() coreiface.Path {
	path, err := coreiface.ParsePath(ipfspath.Join([]string{"/ipns", k.peerID.Pretty()}))
	if err != nil {
		panic("error parsing path: " + err.Error())
	}

	return path
}

//
func (k *key) ID() peer.ID {
	return k.peerID
}

//生成生成新密钥，并将其存储在指定的密钥库中
//名称并返回其公钥的base58编码多哈希。
func (api *KeyAPI) Generate(ctx context.Context, name string, opts ...caopts.KeyGenerateOption) (coreiface.Key, error) {
	options, err := caopts.KeyGenerateOptions(opts...)
	if err != nil {
		return nil, err
	}

	if name == "self" {
		return nil, fmt.Errorf("cannot create key with name 'self'")
	}

	_, err = api.repo.Keystore().Get(name)
	if err == nil {
		return nil, fmt.Errorf("key with name '%s' already exists", name)
	}

	var sk crypto.PrivKey
	var pk crypto.PubKey

	switch options.Algorithm {
	case "rsa":
		if options.Size == -1 {
			options.Size = caopts.DefaultRSALen
		}

		priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, options.Size, rand.Reader)
		if err != nil {
			return nil, err
		}

		sk = priv
		pk = pub
	case "ed25519":
		priv, pub, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return nil, err
		}

		sk = priv
		pk = pub
	default:
		return nil, fmt.Errorf("unrecognized key type: %s", options.Algorithm)
	}

	err = api.repo.Keystore().Put(name, sk)
	if err != nil {
		return nil, err
	}

	pid, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return nil, err
	}

	return &key{name, pid}, nil
}

//list返回存储在keystore中的列表键。
func (api *KeyAPI) List(ctx context.Context) ([]coreiface.Key, error) {
	keys, err := api.repo.Keystore().List()
	if err != nil {
		return nil, err
	}

	sort.Strings(keys)

	out := make([]coreiface.Key, len(keys)+1)
	out[0] = &key{"self", api.identity}

	for n, k := range keys {
		privKey, err := api.repo.Keystore().Get(k)
		if err != nil {
			return nil, err
		}

		pubKey := privKey.GetPublic()

		pid, err := peer.IDFromPublicKey(pubKey)
		if err != nil {
			return nil, err
		}

		out[n+1] = &key{k, pid}
	}
	return out, nil
}

//将“oldname”重命名为“newname”。返回键以及是否另一个键
//密钥被覆盖，或出现错误。
func (api *KeyAPI) Rename(ctx context.Context, oldName string, newName string, opts ...caopts.KeyRenameOption) (coreiface.Key, bool, error) {
	options, err := caopts.KeyRenameOptions(opts...)
	if err != nil {
		return nil, false, err
	}

	ks := api.repo.Keystore()

	if oldName == "self" {
		return nil, false, fmt.Errorf("cannot rename key with name 'self'")
	}

	if newName == "self" {
		return nil, false, fmt.Errorf("cannot overwrite key with name 'self'")
	}

	oldKey, err := ks.Get(oldName)
	if err != nil {
		return nil, false, fmt.Errorf("no key named %s was found", oldName)
	}

	pubKey := oldKey.GetPublic()

	pid, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return nil, false, err
	}

//这很重要，因为将来的代码将删除关键字“oldname”
//即使它与newname相同。
	if newName == oldName {
		return &key{oldName, pid}, false, nil
	}

	overwrite := false
	if options.Force {
		exist, err := ks.Has(newName)
		if err != nil {
			return nil, false, err
		}

		if exist {
			overwrite = true
			err := ks.Delete(newName)
			if err != nil {
				return nil, false, err
			}
		}
	}

	err = ks.Put(newName, oldKey)
	if err != nil {
		return nil, false, err
	}

	return &key{newName, pid}, overwrite, ks.Delete(oldName)
}

//移除从密钥库中移除密钥。返回已删除密钥的IPN路径。
func (api *KeyAPI) Remove(ctx context.Context, name string) (coreiface.Key, error) {
	ks := api.repo.Keystore()

	if name == "self" {
		return nil, fmt.Errorf("cannot remove key with name 'self'")
	}

	removed, err := ks.Get(name)
	if err != nil {
		return nil, fmt.Errorf("no key named %s was found", name)
	}

	pubKey := removed.GetPublic()

	pid, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	err = ks.Delete(name)
	if err != nil {
		return nil, err
	}

	return &key{"", pid}, nil
}

func (api *KeyAPI) Self(ctx context.Context) (coreiface.Key, error) {
	if api.identity == "" {
		return nil, errors.New("identity not loaded")
	}

	return &key{"self", api.identity}, nil
}

