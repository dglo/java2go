package parser

import (
	"fmt"
	"io"
	"log"
	"sort"
)

type MethodMapper interface {
	FixDuplicate(mlist []GoMethod, newmthd GoMethod) bool
}

type MethodMap struct {
	methods map[string][]GoMethod
}

func (mmap *MethodMap) AddMethod(newmthd GoMethod, mapper MethodMapper) {
	if mmap.methods == nil {
		mmap.methods = make(map[string][]GoMethod)
	}
	if mthd, ok := mmap.methods[newmthd.Name()]; !ok {
		mmap.methods[newmthd.Name()] = make([]GoMethod, 1)
		mmap.methods[newmthd.Name()][0] = newmthd
	} else {
		if !mapper.FixDuplicate(mthd, newmthd) {
			mmap.methods[newmthd.Name()] =
				append(mmap.methods[newmthd.Name()], newmthd)
		}
	}
}

func (mmap *MethodMap) FindMethod(name string,
	args *GoMethodArguments) GoMethod {
	for key, mlist := range mmap.methods {
		if key == name {
			for _, m := range mlist {
				if m.HasArguments(args) {
					return m
				}
			}
		}
	}

	return nil
}

func (mmap *MethodMap) Length() int {
	return len(mmap.methods)
}

func (mmap *MethodMap) MethodList(name string) []GoMethod {
	if mmap.methods != nil {
		if mthd, ok := mmap.methods[name]; ok {
			return mthd
		}
	}

	return nil
}

func (mmap *MethodMap) renumberDuplicateMethods(gp *GoProgram) {
	for _, mlist := range mmap.methods {
		if len(mlist) > 1 {
			n := 0
			for _, m := range mlist {
				if n > 0 {
					m.SetGoName(fmt.Sprintf("%s%d", m.GoName(), n + 1))
				}
				n += 1
			}
		}
	}
}

func (mmap *MethodMap) SortedKeys() []string {
	if mmap.methods == nil {
		return nil
	}

	flds := make([]string, len(mmap.methods))
	i := 0
	for k, _ := range mmap.methods {
		flds[i] = k
		i++
	}

	sort.Sort(sort.StringSlice(flds))

	return flds
}

func (mmap *MethodMap) WriteString(out io.Writer, verbose bool) {
	io.WriteString(out, "|")
	if !verbose {
		io.WriteString(out, fmt.Sprintf("%d methods", len(mmap.methods)))
	} else {
		no_comma := true
		for _, mlist := range mmap.methods {
			for _, m := range mlist {
				if no_comma {
					no_comma = false
				} else {
					io.WriteString(out, ",")
				}
				m.WriteString(out)
			}
		}
	}
}

type ClassMethodMap struct {
	MethodMap
}

func NewClassMethodMap() *ClassMethodMap {
	return &ClassMethodMap{MethodMap{}}
}

func (cmm *ClassMethodMap) FixDuplicate(mlist []GoMethod, newmthd GoMethod) bool {
	fixed := false
	if gm, is_gm := newmthd.(*GoClassMethod); !is_gm {
		if _, is_ref := newmthd.(*GoMethodReference); !is_ref {
			log.Printf("//ERR// Unknown class method type %T\n", newmthd)
		}
		fixed = true
	} else {
		for i, m2 := range mlist {
			if m2.IsMethod(newmthd) {
				if _, m2ok := m2.(*GoMethodReference); m2ok {
					m2.SetOriginal(gm)
					mlist[i] = gm
					fixed = true
				}
			}
		}
	}
	return fixed
}

type InterfaceMethodMap struct {
	MethodMap
}

func NewInterfaceMethodMap() *InterfaceMethodMap {
	return &InterfaceMethodMap{MethodMap{}}
}

func (cmm *InterfaceMethodMap) FixDuplicate(mlist []GoMethod, newmthd GoMethod) bool {
	fixed := false
	if gm, is_gm := newmthd.(*GoIfaceMethod); !is_gm {
		if _, is_ref := newmthd.(*GoMethodReference); !is_ref {
			log.Printf("//ERR// Unknown interface method type %T\n", newmthd)
		}
		fixed = true
	} else {
		for i, m2 := range mlist {
			if m2.IsMethod(newmthd) {
				if _, m2ok := m2.(*GoMethodReference); m2ok {
					//m2.SetOriginal(gm)
					log.Printf("//ERR// Not setting original for interface" +
						" method %T\n", newmthd)
					mlist[i] = gm
					fixed = true
				}
			}
		}
	}
	return fixed
}
