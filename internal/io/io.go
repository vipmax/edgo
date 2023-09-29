package io

import (
	"github.com/rjeczalik/notify"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)


type FileWatcher struct {
	filePath  string
	lastStats os.FileInfo
	ticker    *time.Ticker
	mu        sync.Mutex
}

func NewFileWatcher(everyms int) *FileWatcher {
	duration := time.Millisecond * time.Duration(everyms)
	ticker := time.NewTicker(duration)

	return &FileWatcher{
		filePath:  "",
		lastStats: nil,
		ticker:    ticker,
		mu:        sync.Mutex{},
	}
}

func (fw *FileWatcher) StartWatch(onUpdate func() ) {
	go func() {
		for range fw.ticker.C {
			//if fw == nil { continue }
			//if fw.lastStats == nil { continue }
			if fw.filePath == "" { continue }
			stats, err := os.Stat(fw.filePath)
			if err != nil { continue }

			if stats.Size() != fw.lastStats.Size() ||
			   stats.ModTime() != fw.lastStats.ModTime() {
				onUpdate()
				fw.UpdateStats()
			}
		}
	}()
}


func (fw *FileWatcher) UpdateFile(filePath string) {
	if fw.filePath == filePath { return }
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.filePath = filePath
}

// invoke only if needs to update stats manually
func (fw *FileWatcher) UpdateStats() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	newStats, _ := os.Stat(fw.filePath)
	fw.lastStats = newStats
	//fw.ticker.Reset(time.Millisecond * 1000)
}


type DirWatcher struct {
	dirPath  string
	events chan notify.EventInfo
	mu 	  sync.Mutex
}

func NewDirWatcher(dirPath string) *DirWatcher {
	abs, _ := filepath.Abs(dirPath)

	return &DirWatcher{
		dirPath:  abs,
		mu:        sync.Mutex{},
	}
}

func (dw *DirWatcher) StartWatch(onUpdate func(e notify.EventInfo) ) {
	dw.events = make(chan notify.EventInfo, 1)
	path := dw.dirPath + "/..." // any dirs and files inside

	err := notify.Watch(path, dw.events, notify.All)
	if err != nil { log.Fatal(err) }

	go func() {
		for e := range dw.events {
			//log.Println("Got event:", e)
			onUpdate(e)
		}
	}()
}
