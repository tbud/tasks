package nexus

import (
	"fmt"
	. "github.com/tbud/bud/context"
	"github.com/tbud/x/container/set"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type NexusCleanTask struct {
	RepositoryDir string
	JarKeepNum    int
	WarKeepNum    int
	JarKeepDays   int
	WarKeepDays   int
	Test          bool // test or real do remove action
}

var dirSet = set.StringSet{}

func init() {
	nst := &NexusCleanTask{JarKeepNum: 20, WarKeepNum: 20, Test: false}

	Task("clean", Group("nexus"), nst, Usage("Clean nexus repository useless package."))

	nstTest := &NexusCleanTask{JarKeepNum: 20, WarKeepNum: 20, Test: true}
	Task("test", Group("nexus"), nstTest, Usage("Test nexus repository useless package."))
}

func (n *NexusCleanTask) Execute() (err error) {
	reposDir := n.RepositoryDir
	if !filepath.IsAbs(reposDir) {
		if reposDir, err = filepath.Abs(reposDir); err != nil {
			return err
		}
	}

	var (
		lastRootPath string
		lastDirInfo  os.FileInfo
		isWarPath    bool
		vs           versions
	)

	filepath.Walk(reposDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			Log.Error("walk path err: %v", err)
			return nil
		}

		if info.Name() == ".nexus" && info.IsDir() {
			return filepath.SkipDir
		}

		dir := filepath.Dir(path)
		if dirSet.Has(dir) {
			if info.IsDir() {
				vs = append(vs, info)
			}

			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			switch ext {
			case ".jar", ".war":
				dir := filepath.Dir(path)
				dir = filepath.Dir(dir)
				if !dirSet.Has(dir) {
					dirSet.Add(dir)
					// do last root path clean
					if len(vs) > 0 {
						if isWarPath {
							cleanPath(lastRootPath, vs, n.WarKeepNum, n.WarKeepDays, n.Test)
						} else {
							cleanPath(lastRootPath, vs, n.JarKeepNum, n.JarKeepDays, n.Test)
						}
						vs = vs[:0]
					}

					// save last root info
					lastRootPath = dir
					vs = append(vs, lastDirInfo)
					isWarPath = ext == ".war"
				} else {
					if ext == ".war" {
						isWarPath = true
					}
				}
			}
		} else {
			lastDirInfo = info
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

func cleanPath(root string, vs versions, keepNum int, keepDays int, test bool) {
	fmt.Printf("path '%s' version num %d\n", root, len(vs))

	if len(vs) > keepNum {
		sort.Sort(vs)
		keeps := len(vs) - keepNum
		removeVersions := vs[0:keeps]
		keepVersions := vs[keeps:]

		t := time.Now()
		t = t.AddDate(0, 0, -keepDays)

		var keepForDays versions
		for i, r := range removeVersions {
			if r.ModTime().After(t) {
				keepForDays = removeVersions[i:]
				removeVersions = removeVersions[0:i]
				break
			}
		}

		dealWithRemoveVersion(removeVersions, root, test)

		if len(keepForDays) > 0 {
			dealWithKeepVersion(keepForDays, "days")
		}
		dealWithKeepVersion(keepVersions, "num")
	} else {
		if test {
			dealWithKeepVersion(vs, "num")
		}
	}
}

func dealWithRemoveVersion(removeVersions versions, path string, test bool) {
	fmt.Println("Will removed versioins:")
	for _, remove := range removeVersions {
		if test {
			fmt.Printf("%s\t\t%s\n", remove.Name(), remove.ModTime().Format("2006-01-02 15:04:05.999"))
		} else {
			rd := filepath.Join(path, remove.Name())
			fmt.Printf("remove path '%s'\n", rd)
			os.RemoveAll(rd)
		}
	}
}

func dealWithKeepVersion(keepVersions versions, reason string) {
	fmt.Printf("Will keeped for %s:\n", reason)
	for _, keep := range keepVersions {
		fmt.Printf("%s\t\t%s\n", keep.Name(), keep.ModTime().Format("2006-01-02 15:04:05.999"))
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

	if n.JarKeepDays == 0 {
		return fmt.Errorf("jar keep day must large than 0")
	}

	if n.WarKeepDays == 0 {
		return fmt.Errorf("war keep day must large than 0")
	}

	return nil
}
