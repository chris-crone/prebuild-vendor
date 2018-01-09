package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

var tmpl = `package main

import (
%s
)

func main() {
}

`

type pack struct {
	Deps []string
}

func main() {
	args := os.Args[1:]
	paths := args
	if len(args) == 0 {
		paths = []string{"./..."}
	}
	pkgs := map[string]struct{}{}
	listArgs := []string{"list", "-json"}
	listArgs = append(listArgs, paths...)
	out, err := exec.Command("go", listArgs...).Output()
	if err != nil {
		panic(fmt.Sprintf("Unable to list packages: %s\n", err))
	}
	reader := bytes.NewReader(out)
	decoder := json.NewDecoder(reader)
	vStr := "/vendor/"
	for {
		p := pack{}
		err = decoder.Decode(&p)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		for _, d := range p.Deps {
			if strings.Contains(d, "/internal/") || strings.HasSuffix(d, "/internal") {
				continue
			}
			if strings.Contains(d, vStr) {
				d = d[strings.Index(d, vStr)+len(vStr):]
				pkgs[d] = struct{}{}
			}
		}
	}
	l := sort.StringSlice{}
	for p := range pkgs {
		l = append(l, p)
	}
	sort.Sort(l)
	var imports string
	for _, p := range l {
		imports += fmt.Sprintf("\t_ \"%s\"\n", p)
	}
	fmt.Printf(tmpl, imports)
}

