package Geecache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) { // 接口型函数		func -> interface
	return f(key)
}

type Group struct {
	name      string // group 可以视为缓存的命名空间
	getter    Getter // cache miss callback
	mainCache cache
	peers     PeerPicker
}

var (
	mu     sync.Mutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	mu.Lock()
	groups[name] = g
	mu.Unlock()
	return g
}

func GetGroup(name string) *Group {
	mu.Lock()
	g := groups[name]
	mu.Unlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[Geecache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) { // 分布式设计
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			value, err := g.getFromPeer(peer, key)
			if err == nil {
				return value, nil
			}
			log.Printf("[GeeCache] Failed to get from peer: %s", err)
		}
	}
	return g.getLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: CloneBytes(bytes)}
	g.populateCache(key, value) // get data and add to cache
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}
