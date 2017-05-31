package main

import (
	"log"

	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/util/dirwatch"
	"github.com/docopt/docopt-go"

	"github.com/ring00/torrentfs"
)

func updateClient(client *torrent.Client, event *dirwatch.Event) {
	switch event.Change {
	case dirwatch.Added:
		if event.TorrentFilePath != "" {
			_, err := client.AddTorrentFromFile(event.TorrentFilePath)
			if err != nil {
				log.Printf("error adding %s\n", event.TorrentFilePath)
			}
		} else if event.MagnetURI != "" {
			_, err := client.AddMagnet(event.MagnetURI)
			if err != nil {
				log.Printf("error adding %s\n", event.TorrentFilePath)
			}
		}
	case dirwatch.Removed:
		t, ok := client.Torrent(event.InfoHash)
		if ok {
			// No support yet
			t.Drop()
		}
	default:
		log.Println("Unknown operation.")
	}
}

func main() {
	usage := `BitTorrent Filesystem.

Usage:
  torrentfs --mount <mount> --torrent <torrent> [--download <download>]
  torrentfs -h | --help
  torrentfs --version

Options:
  --mount     Mount point.
  --torrent   Directory containing torrents.
  --download  Directory for storing downloaded files.
  -h --help   Show this screen.
  --version   Show version.`

	// Parsing command line arguments
	arguments, err := docopt.Parse(usage, nil, true, "0.1", false)

	if err != nil {
		panic(err)
	}

	usr, err := user.Current()

	mountDir := arguments["<mount>"].(string)
	torrentDir := arguments["<torrent>"].(string)
	downloadDir := filepath.Join(usr.HomeDir, "Downloads")

	if arguments["<download>"] != nil {
		downloadDir = arguments["download"].(string)
	}

	// Setting up the BitTorrent client
	client, err := torrent.NewClient(&torrent.Config{
		DataDir:     downloadDir,
		NoUpload:    true,
		DisableIPv6: false,
		Debug:       true,
	})
	if err != nil {
		panic(err)
	}

	watcher, err := dirwatch.New(torrentDir)
	if err != nil {
		panic(err)
	}

	go func() {
		for event := range watcher.Events {
			updateClient(client, &event)
		}
	}()

	// Setting up FUSE
	conn, err := fuse.Mount(mountDir)
	if err != nil {
		panic(err)
	}
	defer fuse.Unmount(mountDir)
	defer conn.Close()

	btfs := torrentfs.New(client)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		for {
			<-c
			btfs.Destroy()
			err := fuse.Unmount(mountDir)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	if err := fs.Serve(conn, btfs); err != nil {
		log.Fatal(err)
	}
	<-conn.Ready
	if err := conn.MountError; err != nil {
		log.Fatal(err)
	}
}
