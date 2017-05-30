package torrentfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"golang.org/x/net/context"
	"os"
)

const (
	readOnly = 0555
)

type TorrentFS struct {
	root *RootNode
}

func (fs *TorrentFS) Root() (Node, error) {
	return *root, nil
}

func (fs *TorrentFS) Destory() {
	for node := fs.root.Child; node != nil; node = node.Sibling {
		node.Torrent.Drop()
	}
}

func New(client *torrent.Client) *TorrentFS {
	root := &RootNode{client, nil}
	for _, torrent := range client.Torrents() {
		node := newNode(torrent)
		node.Sibling = root.Child
		root.Child = node
	}
	return &TorrentFS{root}
}

type RootNode struct {
	Client *torrent.Client
	Child  *Node
}

func (root *RootNode) Update() {
	for _, torrent := range root.Client.Torrents() {
		update := true
		for node := root.Child; node != nil; node = node.Sibling {
			if node.Name == torrent.Name() {
				update = false
				break
			}
		}

		if update {
			added := newNode(torrent)
			added.Sibling = root.Child
			root.Child = added
		}
	}
}

func (root *RootNode) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir
	return nil
}

func (root *RootNode) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for node := root.Child; node != nil; node = node.Sibling {
		if node.Name == name {
			return *node, nil
		}
	}
	return Node{}, fuse.ENOENT
}

func (root *RootNode) ReadDirAll(ctx context.Context) (entry []fuse.Dirent, err error) {
	root.Update()
	for node := root.Child; node != nil; node = node.Sibling {
		entry = append(entry, fuse.Dirent{Type: node.Type, Name: node.Name})
	}
	return entry, err
}

type Node struct {
	// left-child right-sibling binary tree
	Child, Sibling *Node
	// The torrent of this node
	Torrent *torrent.Torrent
	// Type of this node, may be fuse.DT_File or fuse.DT_Dir
	Type fuse.DirentType
	// Name of this node
	Name string
	// Size of this node
	Length int64
	// Offset of the corresponding content in .torrent
	Offset int64
}

func getTypeFromPath(path []string) fuse.DirentType {
	if len(path) == 1 {
		return fuse.DT_File
	} else {
		return fuse.DT_Dir
	}
}

func (node *Node) Insert(path []string, length int64, offset int64) {
	var root *Node = nil

	for child := node.Child; child != nil; child = child.Sibling {
		if child.Name == path[0] {
			root = child
			break
		}
	}

	if root == nil {
		root = &Node{
			Sibling: node.Child,
			Torrent: node.Torrent,
			Type:    getTypeFromPath(path),
			Name:    path[0],
			Length:  length,
			Offset:  offset,
		}
		node.Child = root
	} else {
		root.Length += length
	}

	if len(path) > 1 {
		root.Insert(path[1:], length, offset)
	}
}

func (node *Node) Build(info *metainfo.Info) {
	for _, finfo := range info.Files {
		node.Insert(finfo.Path, finfo.Length, finfo.Offset(info))
	}
}

func (node *Node) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Size = node.Length
	attr.Mode = readOnly
	if node.Type == fuse.DT_Dir {
		attr.Mode |= os.ModeDir
	}
	return nil
}

func (node *Node) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for child := root.Child; child != nil; child = child.Sibling {
		if child.Name == name {
			return *child, nil
		}
	}
	return Node{}, fuse.ENOENT
}

func (node *Node) ReadDirAll(ctx context.Context) (entry []fuse.Dirent, err error) {
	for child := root.Child; child != nil; child = child.Sibling {
		entry = append(entry, fuse.Dirent{Type: child.Type, Name: child.Name})
	}
	return entry, err
}

func (node *Node) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if node.Type == fuse.DT_Dir {
		panic("Reading from a directory!")
	}

	var err error
	done := make(chan struct{})

	go func() {
		defer close(done)

		reader := node.Torrent.NewReader()
		defer reader.Close()

		_, err := reader.Seek(node.Offset+req.Offset, os.SEEK_SET)
		if err != nil {
			return
		}

		_, err = reader.Read(resp.Data[:req.Size])
	}()

	select {
	case <-done:
		return err
	case <-ctx.Done():
		return fuse.EINTR
	}
}

func newNode(torrent *torrent.Torrent) *Node {
	info := torrent.Info()
	node := &Node{
		Torrent: torrent,
		Type:    fuse.DT_Dir,
		Name:    torrent.Name(),
		Length:  torrent.Length(),
	}
	node.Build(info)
	return node
}
