package Geecache

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 选择节点
}

type PeerGetter interface {
	Get(group string, key string) ([]byte, error) // 查找缓存值
}
