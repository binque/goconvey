package watcher

import (
	"errors"
	"fmt"
	"github.com/smartystreets/goconvey/web/server/contract"
	"log"
	"os"
	"strings"
)

type Watcher struct {
	fs      contract.FileSystem
	shell   contract.Shell
	watched map[string]*contract.Package
	root    string
}

func (self *Watcher) Root() string {
	return self.root
}

func (self *Watcher) Adjust(root string) error {
	if !self.fs.Exists(root) {
		return errors.New(fmt.Sprintf("Directory does not exist: '%s'", root))
	}
	log.Println("Adjusting to watch new root:", self.root)

	self.root = root
	self.watched = make(map[string]*contract.Package)
	self.fs.Walk(root, self.includeFolders)
	self.setGoPath(root)

	return nil
}
func (self *Watcher) includeFolders(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		log.Println("Including:", path)
		self.watched[path] = contract.NewPackage(path)
	}
	return nil
}
func (self *Watcher) setGoPath(root string) {
	if rootGoPathEnd := strings.Index(root, "/src"); rootGoPathEnd >= 0 {
		self.shell.Setenv("GOPATH", root[:rootGoPathEnd])
	} else {
		self.shell.Setenv("GOPATH", root)
	}
}

func (self *Watcher) Deletion(folder string) {
	log.Println("Detected deletion of:", folder)
	delete(self.watched, folder)
}

func (self *Watcher) Creation(folder string) {
	log.Println("Detected creation of:", folder)
	self.watched[folder] = contract.NewPackage(folder)
}

func (self *Watcher) Ignore(packageName string) {
	for key, value := range self.watched {
		if strings.HasSuffix(key, packageName) {
			value.Active = false
		}
	}
}
func (self *Watcher) Reinstate(packageName string) {
	for key, value := range self.watched {
		if strings.HasSuffix(key, packageName) {
			value.Active = true
		}
	}
}
func (self *Watcher) WatchedFolders() []*contract.Package {
	i, watched := 0, make([]*contract.Package, len(self.watched))
	log.Println("Number of watched folders:", len(self.watched))
	for _, item := range self.watched {
		watched[i] = &contract.Package{
			Active: item.Active,
			Path:   item.Path,
			Name:   item.Name,
			Result: contract.NewPackageResult(item.Name),
		}
		i++
	}
	return watched
}

func (self *Watcher) IsWatched(folder string) bool {
	if value, exists := self.watched[folder]; exists {
		return value.Active
	}
	return false
}

func (self *Watcher) IsIgnored(folder string) bool {
	if value, exists := self.watched[folder]; exists {
		return !value.Active
	}
	return false
}

func NewWatcher(fs contract.FileSystem, shell contract.Shell) *Watcher {
	self := &Watcher{}
	self.fs = fs
	self.shell = shell
	self.watched = map[string]*contract.Package{}
	return self
}
