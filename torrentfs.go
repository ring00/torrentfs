package torrentfs

import (
    "os"
    "fmt"
    "path/filepath"
    "strings"
    fusefs "bazil.org/fuse/fs"
    "github.com/anacrolix/torrent"
    "github.com/anacrolix/torrent/metainfo"
)

type TorrentFS struct {
    client *torrent.Client
}

// Root returns the Node for the file system root.
func (fs TorrentFS) Root() (Node, error) {
    return Root{&fs}, nil
}

// Destroy cleans up the file system when it's shutting down.
func (fs TorrentFS) Destory() {

}

// Root implements both Node and Handle for the root directory.
type Root struct {
    fs *TorrentFS
}

type Node struct {
    path []string
    info *metainfo.Info
    fs *TorrentFS
    torrent *torrent.Torrent
}

type Dir struct {
    Node
}

type File struct {
    Node
    size int64
    offset int64
}

// Attr fills attr with the standard metadata for the node.
func (root Root) Attr(ctx context.Context, attr *fuse.Attr) error {
    attr.Mode = os.ModeDir
    return nil
}

// Lookup looks up a specific entry in the receiver,
// which must be a directory.  Lookup should return a Node
// corresponding to the entry.  If the name does not exist in
// the directory, Lookup should return ENOENT.
func (root Root) Lookup(ctx context.Context, name string) (node fusefs.Node, err error) {
    for _, torrent := range root.fs.client.Torrents() {
        info = torrent.Info()
        if info == nil || torrent.Name() != name {
            continue
        }

        _node := Node{make([]string), info, root.fs, torrent}

        if info.IsDir() {
            node = Dir{_node}
        } else {
            node = File{_node, int64(info.Length), 0}
        }
        break
    }
    if node == nil {
        err = fuse.ENOENT
    }
    return
}

// ReadDirAll returns all Dirent under the root directory.
func (root Root) ReadDirAll(ctx context.Context) (dirents []fuse.Dirent, err error) {
    for _, torrent := range root.fs.client.Torrents() {
        info = torrent.Info()
        if info == nil {
            continue
        }

        var _type fuse.DirentType
        if info.IsDir() {
            _type = fuse.DT_Dir
        } else {
            _type = fuse.DT_File
        }

        dirents = append(dirents, fuse.Dirent{info.Name, _type})
    }
    return
}

func (root Root) Forget() {
    rn.fs.Destroy()
}

func (dir Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
    attr.Mode = os.ModeDir | 0555
    return nil
}

func (dir Dir) Lookup(ctx context.Context, name string) (node fusefs.Node, err error) {

}

func isSubPath(path, basepath string) bool {
    if len(basepath) == 0 {
        return len(path) > 0
    }

    if len(path) <= len(basepath) {
        return false
    }

    if !strings.HasPrefix(path, basepath) {
        return false
    }

    return path[len(basepath)] == '/'
}

func (dir Dir) ReadDirAll(ctx context.Context) (dirents []fuse.Dirent, err error) {
    visited := make(map[string]bool)
    for _, file := range dir.info.Files {
        if !isSubPath(strings.Join(file.Path, "/"), strings.Join(dir.path, "/")) {
            continue
        }
        name = file.Path[len(dir.path)]
        if visited[name] {
            continue
        }

        visited[name] = true
        var entryType fuse.DirentType;
        if info.IsDir() {
            entryType = fuse.DT_Dir
        } else {
            entryType = fuse.DT_File
        }
    }
}

func (file File) Attr(ctx context.Context, attr *fuse.Attr) error {

}

func (file File) Lookup(ctx context.Context, name string) (node fusefs.Node, err error) {

}
