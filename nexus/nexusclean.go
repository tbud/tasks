package nexus

import (
	"fmt"
	. "github.com/tbud/bud/context"
	// . "github.com/tbud/x/config"
	"github.com/tbud/x/container/set"
	"os"
	"path/filepath"
	"sort"
)

type NexusCleanTask struct {
	RepositoryDir string
	JarKeepNum    int
	WarKeepNum    int
	Test          bool
}

var dirSet = set.StringSet{}

func init() {
	nst := &NexusCleanTask{JarKeepNum: 20, WarKeepNum: 20, Test: true}

	Task("clean", Group("nexus"), nst, Usage("Clean nexus repository useless package."))

	// Task("cleantest", Group("nexus"), Tasks("nexus.clean"), Config{"test": true}, Usage("Test nexus repository useless package."))
}

func (n *NexusCleanTask) Execute() (err error) {
	reposDir := n.RepositoryDir
	if !filepath.IsAbs(reposDir) {
		if reposDir, err = filepath.Abs(reposDir); err != nil {
			return err
		}
	}

	filepath.Walk(reposDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			Log.Error("walk path err: %v", err)
			return nil
		}

		dir := filepath.Dir(path)
		if dirSet.Has(dir) {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			switch ext {
			case ".jar", ".war":
				dir := filepath.Dir(path)
				dir = filepath.Dir(dir)
				dirSet.Add(dir)

				if ext == ".war" {
					cleanPath(dir, n.WarKeepNum, n.Test)
				} else {
					cleanPath(dir, n.JarKeepNum, n.Test)
				}
			}
		}

		return nil
	})

	return nil
}

type versions []os.FileInfo

func (v versions) Len() int {
	return len(v)
}

func (v versions) Less(i, j int) bool {
	return v[i].ModTime().UnixNano() < v[j].ModTime().UnixNano()
}

func (v versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func cleanPath(root string, keepNum int, test bool) {
	v := versions{}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if root == path {
			return nil
		}

		if info.IsDir() {
			v = append(v, info)
		}

		return filepath.SkipDir
	})

	fmt.Printf("path '%s' version num %d\n", root, len(v))

	if len(v) > keepNum {
		sort.Sort(v)
		keeps := len(v) - keepNum
		removeVersions := v[0:keeps]
		keepVersions := v[keeps:]

		dealWithRemoveVersion(removeVersions, root, test)
		dealWithKeepVersion(keepVersions)
	} else {
		dealWithKeepVersion(v)
	}
}

func dealWithRemoveVersion(removeVersions versions, path string, test bool) {
	fmt.Println("Will removed versioins:")
	for _, remove := range removeVersions {
		if test {
			fmt.Printf("%s\n", remove.Name())
		} else {
			rd := filepath.Join(path, remove.Name())
			fmt.Printf("remove path '%s'\n", rd)
			os.RemoveAll(rd)
		}
	}
}

func dealWithKeepVersion(keepVersions versions) {
	fmt.Println("Will keeped versions:")
	for _, keep := range keepVersions {
		fmt.Printf("%s\n", keep.Name())
	}
}

func (n *NexusCleanTask) Validate() error {
	if len(n.RepositoryDir) == 0 {
		return fmt.Errorf("respsitory dir is empty")
	}

	if n.JarKeepNum == 0 {
		return fmt.Errorf("jar keep num must large than 0")
	}

	if n.WarKeepNum == 0 {
		return fmt.Errorf("war keep num must large than 0")
	}

	return nil
}
