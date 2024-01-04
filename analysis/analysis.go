package analysis

import (
	"fmt"
	"go/build"
	"log"
	"strings"
	"time"

	"code.byted.org/gopkg/jsonx"
	analysisV "go.callvis_core/analysisV/analysis"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

//关键函数定义
type Fixed struct {
	FuncDesc
	RelationsTree *MWTNode //反向调用关系，可能有多条调用链到达关键函数
	RelationList  []CalledRelation
	CanFix        bool //能反向找到gin.Context即可以自动修复
}

//函数定义
type FuncDesc struct {
	File    string //文件路径
	Package string //package名
	Name    string //函数名，格式为Package.Func
	//函数声明or调用行数
}

//描述一个函数调用N个函数的一对多关系
type CallerRelation struct {
	Caller  FuncDesc
	Callees []FuncDesc
}

//描述关键函数的一条反向调用关系
type CalledRelation struct {
	Callees []FuncDesc
	CanFix  bool //该调用关系能反向找到gin.Context即可以自动修复
}

var Analysis *analysis

type analysis struct {
	prog   *ssa.Program
	pkgs   []*ssa.Package
	mains  []*ssa.Package
	result *pointer.Result
}

func DoAnalysis(args []string) *analysis {
	t0 := time.Now()
	ppkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
	}, args...)
	if err != nil {
		panic(err)
	}
	if packages.PrintErrors(ppkgs) > 0 {
		panic("packages.PrintErrors")
	}
	prog, pkgs := ssautil.Packages(ppkgs, 0) //TODO 为啥是0？
	prog.Build()
	var mains []*ssa.Package
	for _, pkg := range pkgs {
		log.Printf("pkg name:%v\n", pkg.Pkg.Name())
	}
	mains = append(mains, ssautil.MainPackages(pkgs)...)
	if len(mains) == 0 {
		panic("no main packages")
	}
	log.Printf("building.. %d packages (%d main) took: %v",
		len(pkgs), len(mains), time.Since(t0))

	t0 = time.Now()
	ptrcfg := &pointer.Config{
		Mains:          mains,
		BuildCallGraph: true,
	}
	result, err := pointer.Analyze(ptrcfg)
	if err != nil {
		errMsg := fmt.Sprintf("analyze failed:%v", err)
		panic(errMsg)
	}
	log.Printf("analysis took: %v", time.Since(t0))

	return &analysis{
		prog:   prog,
		pkgs:   pkgs,
		mains:  mains,
		result: result,
	}
}

type renderOpts struct {
	focus   string
	group   []string
	ignore  []string
	include []string
	limit   []string
	nointer bool
	nostd   bool
}

func (a *analysis) Render(project string) (map[string]CallerRelation, error) {
	analysisV.P()
	var err error
	var focusPkg *build.Package
	opts := renderOpts{
		//focus: focus,
		//group:  []string{"controller"},
		//ignore: []string{"third", "backend/common", fmt.Sprintf("%s/vendor", project)},
		//include: []string{"backend/code_inspector/testing_bai"},
		//limit: []string{"backend/code_inspector"},
		//nointer: nointer,
		nostd: true,
	}

	callMap, err := printOutput(a.prog, a.mains[0].Pkg, a.result.CallGraph,
		focusPkg, opts.limit, opts.ignore, opts.include, opts.group, opts.nostd, opts.nointer)
	if err != nil {
		return nil, fmt.Errorf("processing failed: %v", err)
	}

	return callMap, nil
}

func traverTreeFromArg(traver string) *MWTNode {
	nodes := strings.Split(traver, "|")
	if len(nodes) != 2 {
		log.Printf("[TraverTreeFromArg] nodes is illegal:%v", nodes)
		return nil
	}
	return &MWTNode{
		Key: fmt.Sprintf("%v.%v", nodes[1], nodes[2]),
		Value: FuncDesc{
			File:    fmt.Sprintf("%v.%v", nodes[1], nodes[2]),
			Package: nodes[1],
			Name:    nodes[2],
		},
		Children: make([]*MWTNode, 0),
	}
}

func (a *analysis) PrintOutput(callMap map[string]CallerRelation, traver string) {
	for k, v := range callMap {
		log.Printf("正向调用关系:%s %+v", k, v)
	}
	tree := traverTreeFromArg(traver)
	if tree != nil {
		log.Printf("逆向调用关系tree:%v", jsonx.ToString(tree))
		BuildFromCallMap(tree, callMap)
		re := CalledRelation{
			Callees: make([]FuncDesc, 0),
		}
		list := make([]CalledRelation, 0)
		depthTraversal(tree, "", re, &list)
		for i := range list {
			log.Printf("list%d: %+v", i, list[i])
		}
	}
}
