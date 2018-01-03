// Copyright (c) 2016, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func prepareTest(c *testCase) {
	c.common = fullProg(c.common)
	c.bash = fullProg(c.bash)
	c.posix = fullProg(c.posix)
	c.mksh = fullProg(c.mksh)
	c.bsmk = fullProg(c.bsmk) // bash AND mksh
	if f, ok := c.common.(*File); ok && f != nil {
		c.All = append(c.All, f)
		c.Bash = f
		c.Posix = f
		c.MirBSDKorn = f
	}
	if f, ok := c.bash.(*File); ok && f != nil {
		c.All = append(c.All, f)
		c.Bash = f
	}
	if f, ok := c.posix.(*File); ok && f != nil {
		c.All = append(c.All, f)
		c.Posix = f
	}
	if f, ok := c.mksh.(*File); ok && f != nil {
		c.All = append(c.All, f)
		c.MirBSDKorn = f
	}
	if f, ok := c.bsmk.(*File); ok && f != nil {
		c.All = append(c.All, f)
		c.Bash = f
		c.MirBSDKorn = f
	}
}

func init() {
	for i := range fileTests {
		prepareTest(&fileTests[i])
	}
	for i := range fileTestsNoPrint {
		prepareTest(&fileTestsNoPrint[i])
	}
}

func lit(s string) *Lit         { return &Lit{Value: s} }
func word(ps ...WordPart) *Word { return &Word{Parts: ps} }
func litWord(s string) *Word    { return word(lit(s)) }
func litWords(strs ...string) []*Word {
	l := make([]*Word, 0, len(strs))
	for _, s := range strs {
		l = append(l, litWord(s))
	}
	return l
}

func call(words ...*Word) *CallExpr    { return &CallExpr{Args: words} }
func litCall(strs ...string) *CallExpr { return call(litWords(strs...)...) }

func stmt(cmd Command) *Stmt         { return &Stmt{Cmd: cmd} }
func stmtList(sts ...*Stmt) StmtList { return StmtList{Stmts: sts} }
func stmts(cmds ...Command) StmtList {
	l := make([]*Stmt, len(cmds))
	for i, cmd := range cmds {
		l[i] = stmt(cmd)
	}
	return stmtList(l...)
}

func litStmt(strs ...string) *Stmt { return stmt(litCall(strs...)) }
func litStmts(strs ...string) StmtList {
	l := make([]*Stmt, len(strs))
	for i, s := range strs {
		l[i] = litStmt(s)
	}
	return stmtList(l...)
}

func sglQuoted(s string) *SglQuoted        { return &SglQuoted{Value: s} }
func sglDQuoted(s string) *SglQuoted       { return &SglQuoted{Dollar: true, Value: s} }
func dblQuoted(ps ...WordPart) *DblQuoted  { return &DblQuoted{Parts: ps} }
func dblDQuoted(ps ...WordPart) *DblQuoted { return &DblQuoted{Dollar: true, Parts: ps} }
func block(sts ...*Stmt) *Block            { return &Block{StmtList: stmtList(sts...)} }
func subshell(sts ...*Stmt) *Subshell      { return &Subshell{StmtList: stmtList(sts...)} }
func arithmExp(e ArithmExpr) *ArithmExp    { return &ArithmExp{X: e} }
func arithmExpBr(e ArithmExpr) *ArithmExp  { return &ArithmExp{Bracket: true, X: e} }
func arithmCmd(e ArithmExpr) *ArithmCmd    { return &ArithmCmd{X: e} }
func parenArit(e ArithmExpr) *ParenArithm  { return &ParenArithm{X: e} }
func parenTest(e TestExpr) *ParenTest      { return &ParenTest{X: e} }

func cmdSubst(sts ...*Stmt) *CmdSubst { return &CmdSubst{StmtList: stmtList(sts...)} }
func litParamExp(s string) *ParamExp {
	return &ParamExp{Short: true, Param: lit(s)}
}
func letClause(exps ...ArithmExpr) *LetClause {
	return &LetClause{Exprs: exps}
}

func arrValues(words ...*Word) *ArrayExpr {
	ae := &ArrayExpr{}
	for _, w := range words {
		ae.Elems = append(ae.Elems, &ArrayElem{Value: w})
	}
	return ae
}

type testCase struct {
	Strs        []string
	common      interface{}
	bash, posix interface{}
	bsmk, mksh  interface{}
	All         []*File
	Bash, Posix *File
	MirBSDKorn  *File
}

var fileTests = []testCase{
	{
		Strs:   []string{"", " ", "\t", "\n \n", "\r \r\n"},
		common: &File{},
	},
	{
		Strs:   []string{"", "# foo", "# foo ( bar", "# foo'bar"},
		common: &File{},
	},
	{
		Strs:   []string{"foo", "foo ", " foo", "foo # bar"},
		common: litWord("foo"),
	},
	{
		Strs:   []string{`\`},
		common: litWord(`\`),
	},
	{
		Strs:   []string{`foo\`, "f\\\noo\\"},
		common: litWord(`foo\`),
	},
	{
		Strs:   []string{`foo\a`, "f\\\noo\\a"},
		common: litWord(`foo\a`),
	},
	{
		Strs: []string{
			"foo\nbar",
			"foo; bar;",
			"foo;bar;",
			"\nfoo\nbar\n",
			"foo\r\nbar\r\n",
		},
		common: litStmts("foo", "bar"),
	},
	{
		Strs:   []string{"foo a b", " foo  a  b ", "foo \\\n a b"},
		common: litCall("foo", "a", "b"),
	},
	{
		Strs:   []string{"foobar", "foo\\\nbar", "foo\\\nba\\\nr"},
		common: litWord("foobar"),
	},
	{
		Strs:   []string{"foo", "foo \\\n"},
		common: litWord("foo"),
	},
	{
		Strs:   []string{"foo'bar'"},
		common: word(lit("foo"), sglQuoted("bar")),
	},
	{
		Strs:   []string{"(foo)", "(foo;)", "(\nfoo\n)"},
		common: subshell(litStmt("foo")),
	},
	{
		Strs:   []string{"(\n\tfoo\n\tbar\n)", "(foo; bar)"},
		common: subshell(litStmt("foo"), litStmt("bar")),
	},
	{
		Strs:   []string{"{ foo; }", "{\nfoo\n}"},
		common: block(litStmt("foo")),
	},
	{
		Strs: []string{
			"{ if a; then b; fi; }",
			"{ if a; then b; fi }",
		},
		common: block(stmt(&IfClause{
			Cond: litStmts("a"),
			Then: litStmts("b"),
		})),
	},
	{
		Strs: []string{
			"if a; then b; fi",
			"if a\nthen\nb\nfi",
			"if a \nthen\nb\nfi",
		},
		common: &IfClause{
			Cond: litStmts("a"),
			Then: litStmts("b"),
		},
	},
	{
		Strs: []string{
			"if a; then b; else c; fi",
			"if a\nthen b\nelse\nc\nfi",
		},
		common: &IfClause{
			Cond: litStmts("a"),
			Then: litStmts("b"),
			Else: litStmts("c"),
		},
	},
	{
		Strs: []string{
			"if a; then a; elif b; then b; else c; fi",
		},
		common: &IfClause{
			Cond: litStmts("a"),
			Then: litStmts("a"),
			Else: stmtList(stmt(&IfClause{
				Elif: true,
				Cond: litStmts("b"),
				Then: litStmts("b"),
				Else: litStmts("c"),
			})),
		},
	},
	{
		Strs: []string{
			"if a; then a; elif b; then b; elif c; then c; else d; fi",
			"if a\nthen a\nelif b\nthen b\nelif c\nthen c\nelse\nd\nfi",
		},
		common: &IfClause{
			Cond: litStmts("a"),
			Then: litStmts("a"),
			Else: stmtList(stmt(&IfClause{
				Elif: true,
				Cond: litStmts("b"),
				Then: litStmts("b"),
				Else: stmtList(stmt(&IfClause{
					Elif: true,
					Cond: litStmts("c"),
					Then: litStmts("c"),
					Else: litStmts("d"),
				})),
			})),
		},
	},
	{
		Strs: []string{
			"if\n\ta1\n\ta2 foo\n\ta3 bar\nthen b; fi",
			"if a1; a2 foo; a3 bar; then b; fi",
		},
		common: &IfClause{
			Cond: stmtList(
				litStmt("a1"),
				litStmt("a2", "foo"),
				litStmt("a3", "bar"),
			),
			Then: litStmts("b"),
		},
	},
	{
		Strs: []string{`((a == 2))`},
		bsmk: stmt(arithmCmd(&BinaryArithm{
			Op: Eql,
			X:  litWord("a"),
			Y:  litWord("2"),
		})),
		posix: subshell(stmt(subshell(litStmt("a", "==", "2")))),
	},
	{
		Strs: []string{"if (($# > 2)); then b; fi"},
		bsmk: &IfClause{
			Cond: stmts(arithmCmd(&BinaryArithm{
				Op: Gtr,
				X:  word(litParamExp("#")),
				Y:  litWord("2"),
			})),
			Then: litStmts("b"),
		},
	},
	{
		Strs: []string{"(($(date -u) > DATE))"},
		bsmk: arithmCmd(&BinaryArithm{
			Op: Gtr,
			X:  word(cmdSubst(litStmt("date", "-u"))),
			Y:  litWord("DATE"),
		}),
	},
	{
		Strs: []string{": $((0x$foo == 10))"},
		common: call(
			litWord(":"),
			word(arithmExp(&BinaryArithm{
				Op: Eql,
				X:  word(lit("0x"), litParamExp("foo")),
				Y:  litWord("10"),
			})),
		),
	},
	{
		Strs: []string{"((# 1 + 2))", "(( # 1 + 2 ))"},
		mksh: &ArithmCmd{
			X: &BinaryArithm{
				Op: Add,
				X:  litWord("1"),
				Y:  litWord("2"),
			},
			Unsigned: true,
		},
	},
	{
		Strs: []string{"$((# 1 + 2))", "$(( # 1 + 2 ))"},
		mksh: &ArithmExp{
			X: &BinaryArithm{
				Op: Add,
				X:  litWord("1"),
				Y:  litWord("2"),
			},
			Unsigned: true,
		},
	},
	{
		Strs: []string{"((3#20))"},
		bsmk: arithmCmd(litWord("3#20")),
	},
	{
		Strs: []string{
			"while a; do b; done",
			"wh\\\nile a; do b; done",
			"while a\ndo\nb\ndone",
		},
		common: &WhileClause{
			Cond: litStmts("a"),
			Do:   litStmts("b"),
		},
	},
	{
		Strs: []string{"while { a; }; do b; done", "while { a; } do b; done"},
		common: &WhileClause{
			Cond: stmts(block(litStmt("a"))),
			Do:   litStmts("b"),
		},
	},
	{
		Strs: []string{"while (a); do b; done", "while (a) do b; done"},
		common: &WhileClause{
			Cond: stmts(subshell(litStmt("a"))),
			Do:   litStmts("b"),
		},
	},
	{
		Strs: []string{"while ((1 > 2)); do b; done"},
		bsmk: &WhileClause{
			Cond: stmts(arithmCmd(&BinaryArithm{
				Op: Gtr,
				X:  litWord("1"),
				Y:  litWord("2"),
			})),
			Do: litStmts("b"),
		},
	},
	{
		Strs: []string{"until a; do b; done", "until a\ndo\nb\ndone"},
		common: &WhileClause{
			Until: true,
			Cond:  litStmts("a"),
			Do:    litStmts("b"),
		},
	},
	{
		Strs: []string{
			"for i; do foo; done",
			"for i in; do foo; done",
		},
		common: &ForClause{
			Loop: &WordIter{Name: lit("i")},
			Do:   litStmts("foo"),
		},
	},
	{
		Strs: []string{
			"for i in 1 2 3; do echo $i; done",
			"for i in 1 2 3\ndo echo $i\ndone",
			"for i in 1 2 3 #foo\ndo echo $i\ndone",
		},
		common: &ForClause{
			Loop: &WordIter{
				Name:  lit("i"),
				Items: litWords("1", "2", "3"),
			},
			Do: stmts(call(
				litWord("echo"),
				word(litParamExp("i")),
			)),
		},
	},
	{
		Strs: []string{
			"for ((i = 0; i < 10; i++)); do echo $i; done",
			"for ((i=0;i<10;i++)) do echo $i; done",
			"for (( i = 0 ; i < 10 ; i++ ))\ndo echo $i\ndone",
		},
		bash: &ForClause{
			Loop: &CStyleLoop{
				Init: &BinaryArithm{
					Op: Assgn,
					X:  litWord("i"),
					Y:  litWord("0"),
				},
				Cond: &BinaryArithm{
					Op: Lss,
					X:  litWord("i"),
					Y:  litWord("10"),
				},
				Post: &UnaryArithm{
					Op:   Inc,
					Post: true,
					X:    litWord("i"),
				},
			},
			Do: stmts(call(
				litWord("echo"),
				word(litParamExp("i")),
			)),
		},
	},
	{
		Strs: []string{
			"for (( ; ; )); do foo; done",
			"for ((;;)); do foo; done",
		},
		bash: &ForClause{
			Loop: &CStyleLoop{},
			Do:   litStmts("foo"),
		},
	},
	{
		Strs: []string{
			"for ((i = 0; ; )); do foo; done",
			"for ((i = 0;;)); do foo; done",
		},
		bash: &ForClause{
			Loop: &CStyleLoop{
				Init: &BinaryArithm{
					Op: Assgn,
					X:  litWord("i"),
					Y:  litWord("0"),
				},
			},
			Do: litStmts("foo"),
		},
	},
	{
		Strs: []string{
			"select i; do foo; done",
			// TODO: bash won't allow this - bug?
			//"select i in; do foo; done",
		},
		bsmk: &ForClause{
			Select: true,
			Loop:   &WordIter{Name: lit("i")},
			Do:     litStmts("foo"),
		},
	},
	{
		Strs: []string{
			"select i in 1 2 3; do echo $i; done",
			"select i in 1 2 3\ndo echo $i\ndone",
			"select i in 1 2 3 #foo\ndo echo $i\ndone",
		},
		bsmk: &ForClause{
			Select: true,
			Loop: &WordIter{
				Name:  lit("i"),
				Items: litWords("1", "2", "3"),
			},
			Do: stmts(call(
				litWord("echo"),
				word(litParamExp("i")),
			)),
		},
	},
	{
		Strs:  []string{"select foo bar"},
		posix: litStmt("select", "foo", "bar"),
	},
	{
		Strs: []string{`' ' "foo bar"`},
		common: call(
			word(sglQuoted(" ")),
			word(dblQuoted(lit("foo bar"))),
		),
	},
	{
		Strs:   []string{`"foo \" bar"`},
		common: word(dblQuoted(lit(`foo \" bar`))),
	},
	{
		Strs: []string{"\">foo\" \"\nbar\""},
		common: call(
			word(dblQuoted(lit(">foo"))),
			word(dblQuoted(lit("\nbar"))),
		),
	},
	{
		Strs:   []string{`foo \" bar`},
		common: litCall(`foo`, `\"`, `bar`),
	},
	{
		Strs:   []string{`'"'`},
		common: sglQuoted(`"`),
	},
	{
		Strs:   []string{"'`'"},
		common: sglQuoted("`"),
	},
	{
		Strs:   []string{`"'"`},
		common: dblQuoted(lit("'")),
	},
	{
		Strs:   []string{`""`},
		common: dblQuoted(),
	},
	{
		Strs:   []string{"=a s{s s=s"},
		common: litCall("=a", "s{s", "s=s"),
	},
	{
		Strs: []string{"foo && bar", "foo&&bar", "foo &&\nbar"},
		common: &BinaryCmd{
			Op: AndStmt,
			X:  litStmt("foo"),
			Y:  litStmt("bar"),
		},
	},
	{
		Strs: []string{"foo &&\n\tbar"},
		common: &BinaryCmd{
			Op: AndStmt,
			X:  litStmt("foo"),
			Y:  litStmt("bar"),
		},
	},
	{
		Strs: []string{"foo || bar", "foo||bar", "foo ||\nbar"},
		common: &BinaryCmd{
			Op: OrStmt,
			X:  litStmt("foo"),
			Y:  litStmt("bar"),
		},
	},
	{
		Strs: []string{"if a; then b; fi || while a; do b; done"},
		common: &BinaryCmd{
			Op: OrStmt,
			X: stmt(&IfClause{
				Cond: litStmts("a"),
				Then: litStmts("b"),
			}),
			Y: stmt(&WhileClause{
				Cond: litStmts("a"),
				Do:   litStmts("b"),
			}),
		},
	},
	{
		Strs: []string{"foo && bar1 || bar2"},
		common: &BinaryCmd{
			Op: OrStmt,
			X: stmt(&BinaryCmd{
				Op: AndStmt,
				X:  litStmt("foo"),
				Y:  litStmt("bar1"),
			}),
			Y: litStmt("bar2"),
		},
	},
	{
		Strs: []string{"a || b || c || d"},
		common: &BinaryCmd{
			Op: OrStmt,
			X: stmt(&BinaryCmd{
				Op: OrStmt,
				X: stmt(&BinaryCmd{
					Op: OrStmt,
					X:  litStmt("a"),
					Y:  litStmt("b"),
				}),
				Y: litStmt("c"),
			}),
			Y: litStmt("d"),
		},
	},
	{
		Strs: []string{"foo | bar", "foo|bar", "foo |\n#etc\nbar"},
		common: &BinaryCmd{
			Op: Pipe,
			X:  litStmt("foo"),
			Y:  litStmt("bar"),
		},
	},
	{
		Strs: []string{"foo | bar | extra"},
		common: &BinaryCmd{
			Op: Pipe,
			X:  litStmt("foo"),
			Y: stmt(&BinaryCmd{
				Op: Pipe,
				X:  litStmt("bar"),
				Y:  litStmt("extra"),
			}),
		},
	},
	{
		Strs: []string{"foo | a=b bar"},
		common: &BinaryCmd{
			Op: Pipe,
			X:  litStmt("foo"),
			Y: stmt(&CallExpr{
				Assigns: []*Assign{{
					Name:  lit("a"),
					Value: litWord("b"),
				}},
				Args: litWords("bar"),
			}),
		},
	},
	{
		Strs: []string{"foo |&"},
		mksh: &Stmt{Cmd: litCall("foo"), Coprocess: true},
	},
	{
		Strs: []string{"foo |& bar", "foo|&bar"},
		bash: &BinaryCmd{
			Op: PipeAll,
			X:  litStmt("foo"),
			Y:  litStmt("bar"),
		},
		mksh: []*Stmt{
			{Cmd: litCall("foo"), Coprocess: true},
			litStmt("bar"),
		},
	},
	{
		Strs: []string{
			"foo() {\n\ta\n\tb\n}",
			"foo() { a; b; }",
			"foo ( ) {\na\nb\n}",
		},
		common: &FuncDecl{
			Name: lit("foo"),
			Body: stmt(block(litStmt("a"), litStmt("b"))),
		},
	},
	{
		Strs: []string{"foo() { a; }\nbar", "foo() {\na\n}; bar"},
		common: []Command{
			&FuncDecl{
				Name: lit("foo"),
				Body: stmt(block(litStmt("a"))),
			},
			litCall("bar"),
		},
	},
	{
		Strs: []string{"foO_123() { a; }"},
		common: &FuncDecl{
			Name: lit("foO_123"),
			Body: stmt(block(litStmt("a"))),
		},
	},
	{
		Strs: []string{"-foo_.,+-bar() { a; }"},
		bsmk: &FuncDecl{
			Name: lit("-foo_.,+-bar"),
			Body: stmt(block(litStmt("a"))),
		},
	},
	{
		Strs: []string{
			"function foo() {\n\ta\n\tb\n}",
			"function foo {\n\ta\n\tb\n}",
			"function foo() { a; b; }",
		},
		bsmk: &FuncDecl{
			RsrvWord: true,
			Name:     lit("foo"),
			Body:     stmt(block(litStmt("a"), litStmt("b"))),
		},
	},
	{
		Strs: []string{"function foo() (a)"},
		bash: &FuncDecl{
			RsrvWord: true,
			Name:     lit("foo"),
			Body:     stmt(subshell(litStmt("a"))),
		},
	},
	{
		Strs: []string{"a=b foo=$bar foo=start$bar"},
		common: &CallExpr{
			Assigns: []*Assign{
				{Name: lit("a"), Value: litWord("b")},
				{Name: lit("foo"), Value: word(litParamExp("bar"))},
				{Name: lit("foo"), Value: word(
					lit("start"),
					litParamExp("bar"),
				)},
			},
		},
	},
	{
		Strs: []string{"a=\"\nbar\""},
		common: &CallExpr{
			Assigns: []*Assign{{
				Name:  lit("a"),
				Value: word(dblQuoted(lit("\nbar"))),
			}},
		},
	},
	{
		Strs: []string{"A_3a= foo"},
		common: &CallExpr{
			Assigns: []*Assign{{Name: lit("A_3a")}},
			Args:    litWords("foo"),
		},
	},
	{
		Strs: []string{"a=b=c"},
		common: &CallExpr{
			Assigns: []*Assign{{Name: lit("a"), Value: litWord("b=c")}},
		},
	},
	{
		Strs:   []string{"à=b foo"},
		common: litStmt("à=b", "foo"),
	},
	{
		Strs: []string{
			"foo >a >>b <c",
			"foo > a >> b < c",
			">a >>b <c foo",
		},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("a")},
				{Op: AppOut, Word: litWord("b")},
				{Op: RdrIn, Word: litWord("c")},
			},
		},
	},
	{
		Strs: []string{
			"foo bar >a",
			"foo >a bar",
		},
		common: &Stmt{
			Cmd: litCall("foo", "bar"),
			Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("a")},
			},
		},
	},
	{
		Strs: []string{`>a >\b`},
		common: &Stmt{
			Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("a")},
				{Op: RdrOut, Word: litWord(`\b`)},
			},
		},
	},
	{
		Strs: []string{">a\n>b", ">a; >b"},
		common: []*Stmt{
			{Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("a")},
			}},
			{Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("b")},
			}},
		},
	},
	{
		Strs: []string{"foo1\nfoo2 >r2", "foo1; >r2 foo2"},
		common: []*Stmt{
			litStmt("foo1"),
			{
				Cmd: litCall("foo2"),
				Redirs: []*Redirect{
					{Op: RdrOut, Word: litWord("r2")},
				},
			},
		},
	},
	{
		Strs: []string{"foo >bar$(etc)", "foo >b\\\nar`etc`"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{
				{Op: RdrOut, Word: word(
					lit("bar"),
					cmdSubst(litStmt("etc")),
				)},
			},
		},
	},
	{
		Strs: []string{
			"a=b c=d foo >x <y",
			"a=b c=d >x <y foo",
			">x a=b c=d <y foo",
			">x <y a=b c=d foo",
			"a=b >x c=d foo <y",
		},
		common: &Stmt{
			Cmd: &CallExpr{
				Assigns: []*Assign{
					{Name: lit("a"), Value: litWord("b")},
					{Name: lit("c"), Value: litWord("d")},
				},
				Args: litWords("foo"),
			},
			Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("x")},
				{Op: RdrIn, Word: litWord("y")},
			},
		},
	},
	{
		Strs: []string{
			"foo <<EOF\nbar\nEOF",
			"foo <<EOF \nbar\nEOF",
			"foo <<EOF\t\nbar\nEOF",
		},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{"foo <<EOF\n\nbar\nEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("\nbar\n"),
			}},
		},
	},
	{
		Strs: []string{"foo <<EOF\nbar\n\nEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n\n"),
			}},
		},
	},
	{
		Strs: []string{"foo <<EOF\n1\n2\n3\nEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("1\n2\n3\n"),
			}},
		},
	},
	{
		Strs: []string{"a <<EOF\nfoo$bar\nEOF"},
		common: &Stmt{
			Cmd: litCall("a"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: word(
					lit("foo"),
					litParamExp("bar"),
					lit("\n"),
				),
			}},
		},
	},
	{
		Strs: []string{"a <<EOF\n\"$bar\"\nEOF"},
		common: &Stmt{
			Cmd: litCall("a"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: word(
					lit(`"`),
					litParamExp("bar"),
					lit("\"\n"),
				),
			}},
		},
	},
	{
		Strs: []string{"a <<EOF\n$''$bar\nEOF"},
		bash: &Stmt{
			Cmd: litCall("a"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: word(
					lit("$"),
					lit("''"),
					litParamExp("bar"),
					lit("\n"),
				),
			}},
		},
	},
	{
		Strs: []string{
			"a <<EOF\n$(b)\nc\nEOF",
			"a <<EOF\n`b`\nc\nEOF",
		},
		common: &Stmt{
			Cmd: litCall("a"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: word(
					cmdSubst(litStmt("b")),
					lit("\nc\n"),
				),
			}},
		},
	},
	{
		Strs: []string{"a <<EOF\n\\${\nEOF"},
		common: &Stmt{
			Cmd: litCall("a"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("\\${\n"),
			}},
		},
	},
	{
		Strs: []string{
			"{\n\tfoo <<EOF\nbar\nEOF\n}",
			"{ foo <<EOF\nbar\nEOF\n}",
		},
		common: block(&Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n"),
			}},
		}),
	},
	{
		Strs: []string{
			"$(\n\tfoo <<EOF\nbar\nEOF\n)",
			"$(foo <<EOF\nbar\nEOF\n)",
		},
		common: cmdSubst(&Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n"),
			}},
		}),
	},
	{
		Strs: []string{"$(<foo)", "`<foo`"},
		common: cmdSubst(&Stmt{
			Redirs: []*Redirect{{
				Op:   RdrIn,
				Word: litWord("foo"),
			}},
		}),
	},
	{
		Strs: []string{"foo <<EOF >f\nbar\nEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{
				{
					Op:   Hdoc,
					Word: litWord("EOF"),
					Hdoc: litWord("bar\n"),
				},
				{Op: RdrOut, Word: litWord("f")},
			},
		},
	},
	{
		Strs: []string{"foo <<EOF && {\nbar\nEOF\n\tetc\n}"},
		common: &BinaryCmd{
			Op: AndStmt,
			X: &Stmt{
				Cmd: litCall("foo"),
				Redirs: []*Redirect{{
					Op:   Hdoc,
					Word: litWord("EOF"),
					Hdoc: litWord("bar\n"),
				}},
			},
			Y: stmt(block(litStmt("etc"))),
		},
	},
	{
		Strs: []string{
			"$(\n\tfoo\n) <<EOF\nbar\nEOF",
			"<<EOF $(\n\tfoo\n)\nbar\nEOF",
		},
		// note that dash won't accept the second one
		bsmk: &Stmt{
			Cmd: call(word(cmdSubst(litStmt("foo")))),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{
			"$(\n\tfoo\n) <<EOF\nbar\nEOF",
			"`\n\tfoo\n` <<EOF\nbar\nEOF",
			"<<EOF `\n\tfoo\n`\nbar\nEOF",
		},
		common: &Stmt{
			Cmd: call(word(cmdSubst(litStmt("foo")))),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{
			"$((foo)) <<EOF\nbar\nEOF",
			"<<EOF $((\n\tfoo\n))\nbar\nEOF",
		},
		common: &Stmt{
			Cmd: call(word(arithmExp(litWord("foo")))),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{"if true; then foo <<-EOF\n\tbar\n\tEOF\nfi"},
		common: &IfClause{
			Cond: litStmts("true"),
			Then: stmtList(&Stmt{
				Cmd: litCall("foo"),
				Redirs: []*Redirect{{
					Op:   DashHdoc,
					Word: litWord("EOF"),
					Hdoc: litWord("\tbar\n\t"),
				}},
			}),
		},
	},
	{
		Strs: []string{"if true; then foo <<-EOF\n\tEOF\nfi"},
		common: &IfClause{
			Cond: litStmts("true"),
			Then: stmtList(&Stmt{
				Cmd: litCall("foo"),
				Redirs: []*Redirect{{
					Op:   DashHdoc,
					Word: litWord("EOF"),
					Hdoc: litWord("\t"),
				}},
			}),
		},
	},
	{
		Strs: []string{"foo <<EOF\nbar\nEOF\nfoo2"},
		common: []*Stmt{
			{
				Cmd: litCall("foo"),
				Redirs: []*Redirect{{
					Op:   Hdoc,
					Word: litWord("EOF"),
					Hdoc: litWord("bar\n"),
				}},
			},
			litStmt("foo2"),
		},
	},
	{
		Strs: []string{"foo <<FOOBAR\nbar\nFOOBAR"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("FOOBAR"),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{"foo <<\"EOF\"\nbar\nEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: word(dblQuoted(lit("EOF"))),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{"foo <<'EOF'\n${\nEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: word(sglQuoted("EOF")),
				Hdoc: litWord("${\n"),
			}},
		},
	},
	{
		Strs: []string{"foo <<'EOF'\nEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: word(sglQuoted("EOF")),
			}},
		},
	},
	{
		Strs: []string{"foo <<\"EOF\"2\nbar\nEOF2"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: word(dblQuoted(lit("EOF")), lit("2")),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{"foo <<\\EOF\nbar\nEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("\\EOF"),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{
			"foo <<-EOF\nbar\nEOF",
			"foo <<- EOF\nbar\nEOF",
		},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   DashHdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{"foo <<-EOF\n\tEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   DashHdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("\t"),
			}},
		},
	},
	{
		Strs: []string{"foo <<-EOF\n\tbar\n\tEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   DashHdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("\tbar\n\t"),
			}},
		},
	},
	{
		Strs: []string{"foo <<-'EOF'\n\tbar\n\tEOF"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   DashHdoc,
				Word: word(sglQuoted("EOF")),
				Hdoc: litWord("\tbar\n\t"),
			}},
		},
	},
	{
		Strs: []string{
			"f1 <<EOF1\nh1\nEOF1\nf2 <<EOF2\nh2\nEOF2",
			"f1 <<EOF1; f2 <<EOF2\nh1\nEOF1\nh2\nEOF2",
		},
		common: []*Stmt{
			{
				Cmd: litCall("f1"),
				Redirs: []*Redirect{{
					Op:   Hdoc,
					Word: litWord("EOF1"),
					Hdoc: litWord("h1\n"),
				}},
			},
			{
				Cmd: litCall("f2"),
				Redirs: []*Redirect{{
					Op:   Hdoc,
					Word: litWord("EOF2"),
					Hdoc: litWord("h2\n"),
				}},
			},
		},
	},
	{
		Strs: []string{
			"a <<EOF\nfoo\nEOF\nb\nb\nb\nb\nb\nb\nb\nb\nb",
			"a <<EOF;b;b;b;b;b;b;b;b;b\nfoo\nEOF",
		},
		common: []*Stmt{
			{
				Cmd: litCall("a"),
				Redirs: []*Redirect{{
					Op:   Hdoc,
					Word: litWord("EOF"),
					Hdoc: litWord("foo\n"),
				}},
			},
			litStmt("b"), litStmt("b"), litStmt("b"),
			litStmt("b"), litStmt("b"), litStmt("b"),
			litStmt("b"), litStmt("b"), litStmt("b"),
		},
	},
	{
		Strs: []string{
			"foo \"\narg\" <<EOF\nbar\nEOF",
			"foo <<EOF \"\narg\"\nbar\nEOF",
		},
		common: &Stmt{
			Cmd: call(
				litWord("foo"),
				word(dblQuoted(lit("\narg"))),
			),
			Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("bar\n"),
			}},
		},
	},
	{
		Strs: []string{"foo >&2 <&0 2>file <>f2"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{
				{Op: DplOut, Word: litWord("2")},
				{Op: DplIn, Word: litWord("0")},
				{Op: RdrOut, N: lit("2"), Word: litWord("file")},
				{Op: RdrInOut, Word: litWord("f2")},
			},
		},
	},
	{
		Strs: []string{"foo &>a &>>b"},
		bsmk: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{
				{Op: RdrAll, Word: litWord("a")},
				{Op: AppAll, Word: litWord("b")},
			},
		},
		posix: []*Stmt{
			{Cmd: litCall("foo"), Background: true},
			{Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("a")},
			}, Background: true},
			{Redirs: []*Redirect{
				{Op: AppOut, Word: litWord("b")},
			}},
		},
	},
	{
		Strs: []string{"foo 2>file bar", "2>file foo bar"},
		common: &Stmt{
			Cmd: litCall("foo", "bar"),
			Redirs: []*Redirect{
				{Op: RdrOut, N: lit("2"), Word: litWord("file")},
			},
		},
	},
	{
		Strs: []string{"a >f1\nb >f2", "a >f1; b >f2"},
		common: []*Stmt{
			{
				Cmd:    litCall("a"),
				Redirs: []*Redirect{{Op: RdrOut, Word: litWord("f1")}},
			},
			{
				Cmd:    litCall("b"),
				Redirs: []*Redirect{{Op: RdrOut, Word: litWord("f2")}},
			},
		},
	},
	{
		Strs: []string{"foo >|bar"},
		common: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{
				{Op: ClbOut, Word: litWord("bar")},
			},
		},
	},
	{
		Strs: []string{
			"foo <<<input",
			"foo <<< input",
		},
		bsmk: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   WordHdoc,
				Word: litWord("input"),
			}},
		},
	},
	{
		Strs: []string{
			`foo <<<"spaced input"`,
			`foo <<< "spaced input"`,
		},
		bsmk: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op:   WordHdoc,
				Word: word(dblQuoted(lit("spaced input"))),
			}},
		},
	},
	{
		Strs: []string{"foo >(foo)"},
		bash: call(
			litWord("foo"),
			word(&ProcSubst{
				Op:       CmdOut,
				StmtList: litStmts("foo"),
			}),
		),
	},
	{
		Strs: []string{"foo < <(foo)"},
		bash: &Stmt{
			Cmd: litCall("foo"),
			Redirs: []*Redirect{{
				Op: RdrIn,
				Word: word(&ProcSubst{
					Op:       CmdIn,
					StmtList: litStmts("foo"),
				}),
			}},
		},
	},
	{
		Strs: []string{"a<(b) c>(d)"},
		bash: call(
			word(lit("a"), &ProcSubst{
				Op:       CmdIn,
				StmtList: litStmts("b"),
			}),
			word(lit("c"), &ProcSubst{
				Op:       CmdOut,
				StmtList: litStmts("d"),
			}),
		),
	},
	{
		Strs: []string{"! foo"},
		common: &Stmt{
			Negated: true,
			Cmd:     litCall("foo"),
		},
	},
	{
		Strs: []string{"! !"},
		// not common, as dash and mksh reject this (breaking
		// with POSIX)
		bash: &Stmt{
			Negated: true,
			Cmd:     litCall("!"),
		},
	},
	{
		Strs: []string{"foo &\nbar", "foo & bar", "foo&bar"},
		common: []*Stmt{
			{Cmd: litCall("foo"), Background: true},
			litStmt("bar"),
		},
	},
	{
		Strs: []string{"! if foo; then bar; fi >/dev/null &"},
		common: &Stmt{
			Negated: true,
			Cmd: &IfClause{
				Cond: litStmts("foo"),
				Then: litStmts("bar"),
			},
			Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("/dev/null")},
			},
			Background: true,
		},
	},
	{
		Strs:   []string{"foo#bar"},
		common: litWord("foo#bar"),
	},
	{
		Strs:   []string{"{ echo } }; }"},
		common: block(litStmt("echo", "}", "}")),
	},
	{
		Strs: []string{"$({ echo; })"},
		common: cmdSubst(stmt(
			block(litStmt("echo")),
		)),
	},
	{
		Strs: []string{
			"$( (echo foo bar))",
			"$( (echo foo bar) )",
			"`(echo foo bar)`",
		},
		common: cmdSubst(stmt(
			subshell(litStmt("echo", "foo", "bar")),
		)),
	},
	{
		Strs: []string{
			"$(\n\t(a)\n\tb\n)",
			"$( (a); b)",
			"`(a); b`",
		},
		common: cmdSubst(
			stmt(subshell(litStmt("a"))),
			litStmt("b"),
		),
	},
	{
		Strs: []string{
			`$(echo \')`,
			"`" + `echo \\'` + "`",
		},
		common: cmdSubst(litStmt("echo", `\'`)),
	},
	{
		Strs: []string{
			`$(echo \\)`,
			"`" + `echo \\\\` + "`",
		},
		common: cmdSubst(litStmt("echo", `\\`)),
	},
	{
		Strs: []string{"$( (a) | b)"},
		common: cmdSubst(
			stmt(&BinaryCmd{
				Op: Pipe,
				X:  stmt(subshell(litStmt("a"))),
				Y:  litStmt("b"),
			}),
		),
	},
	{
		Strs: []string{`"$( (foo))"`},
		common: dblQuoted(cmdSubst(stmt(
			subshell(litStmt("foo")),
		))),
	},
	{
		Strs: []string{"$({ echo; })", "`{ echo; }`"},
		common: cmdSubst(stmt(
			block(litStmt("echo")),
		)),
	},
	{
		Strs:   []string{`{foo}`},
		common: litWord(`{foo}`),
	},
	{
		Strs:   []string{`{"foo"`},
		common: word(lit("{"), dblQuoted(lit("foo"))),
	},
	{
		Strs:   []string{`foo"bar"`, "fo\\\no\"bar\""},
		common: word(lit("foo"), dblQuoted(lit("bar"))),
	},
	{
		Strs:   []string{`!foo`},
		common: litWord(`!foo`),
	},
	{
		Strs:   []string{"$(foo bar)", "`foo bar`"},
		common: cmdSubst(litStmt("foo", "bar")),
	},
	{
		Strs: []string{"$(foo | bar)", "`foo | bar`"},
		common: cmdSubst(
			stmt(&BinaryCmd{
				Op: Pipe,
				X:  litStmt("foo"),
				Y:  litStmt("bar"),
			}),
		),
	},
	{
		Strs: []string{"$(foo | >f)", "`foo | >f`"},
		common: cmdSubst(
			stmt(&BinaryCmd{
				Op: Pipe,
				X:  litStmt("foo"),
				Y: &Stmt{Redirs: []*Redirect{{
					Op:   RdrOut,
					Word: litWord("f"),
				}}},
			}),
		),
	},
	{
		Strs: []string{"$(foo $(b1 b2))"},
		common: cmdSubst(stmt(call(
			litWord("foo"),
			word(cmdSubst(litStmt("b1", "b2"))),
		))),
	},
	{
		Strs: []string{`"$(foo "bar")"`},
		common: dblQuoted(cmdSubst(stmt(call(
			litWord("foo"),
			word(dblQuoted(lit("bar"))),
		)))),
	},
	{
		Strs:   []string{"$(foo)", "`fo\\\no`"},
		common: cmdSubst(litStmt("foo")),
	},
	{
		Strs: []string{"foo $(bar)", "foo `bar`"},
		common: call(
			litWord("foo"),
			word(cmdSubst(litStmt("bar"))),
		),
	},
	{
		Strs: []string{"$(foo 'bar')", "`foo 'bar'`"},
		common: cmdSubst(stmt(call(
			litWord("foo"),
			word(sglQuoted("bar")),
		))),
	},
	{
		Strs: []string{`$(foo "bar")`, "`foo \"bar\"`"},
		common: cmdSubst(stmt(call(
			litWord("foo"),
			word(dblQuoted(lit("bar"))),
		))),
	},
	{
		Strs: []string{"${ foo;}", "${\n\tfoo; }", "${\tfoo;}"},
		mksh: &CmdSubst{
			StmtList: litStmts("foo"),
			TempFile: true,
		},
	},
	{
		Strs: []string{"${\n\tfoo\n\tbar\n}", "${ foo; bar;}"},
		mksh: &CmdSubst{
			StmtList: litStmts("foo", "bar"),
			TempFile: true,
		},
	},
	{
		Strs: []string{"${|foo;}", "${| foo; }"},
		mksh: &CmdSubst{
			StmtList: litStmts("foo"),
			ReplyVar: true,
		},
	},
	{
		Strs: []string{"${|\n\tfoo\n\tbar\n}", "${|foo; bar;}"},
		mksh: &CmdSubst{
			StmtList: litStmts("foo", "bar"),
			ReplyVar: true,
		},
	},
	{
		Strs:   []string{`"$foo"`},
		common: dblQuoted(litParamExp("foo")),
	},
	{
		Strs:   []string{`"#foo"`},
		common: dblQuoted(lit("#foo")),
	},
	{
		Strs: []string{`$@a $*a $#a $$a $?a $!a $-a $0a $30a $_a`},
		common: call(
			word(litParamExp("@"), lit("a")),
			word(litParamExp("*"), lit("a")),
			word(litParamExp("#"), lit("a")),
			word(litParamExp("$"), lit("a")),
			word(litParamExp("?"), lit("a")),
			word(litParamExp("!"), lit("a")),
			word(litParamExp("-"), lit("a")),
			word(litParamExp("0"), lit("a")),
			word(litParamExp("3"), lit("0a")),
			word(litParamExp("_a")),
		),
	},
	{
		Strs:   []string{`$`, `$ #`},
		common: litWord("$"),
	},
	{
		Strs: []string{`${@} ${*} ${#} ${$} ${?} ${!} ${0} ${-}`},
		common: call(
			word(&ParamExp{Param: lit("@")}),
			word(&ParamExp{Param: lit("*")}),
			word(&ParamExp{Param: lit("#")}),
			word(&ParamExp{Param: lit("$")}),
			word(&ParamExp{Param: lit("?")}),
			word(&ParamExp{Param: lit("!")}),
			word(&ParamExp{Param: lit("0")}),
			word(&ParamExp{Param: lit("-")}),
		),
	},
	{
		Strs: []string{`${#$} ${#@} ${#*} ${##} ${#:-a} ${?+b}`},
		common: call(
			word(&ParamExp{Length: true, Param: lit("$")}),
			word(&ParamExp{Length: true, Param: lit("@")}),
			word(&ParamExp{Length: true, Param: lit("*")}),
			word(&ParamExp{Length: true, Param: lit("#")}),
			word(&ParamExp{Param: lit("#"), Exp: &Expansion{
				Op:   SubstColMinus,
				Word: litWord("a"),
			}}),
			word(&ParamExp{
				Param: lit("?"),
				Exp: &Expansion{
					Op:   SubstPlus,
					Word: litWord("b"),
				},
			}),
		),
	},
	{
		Strs:   []string{`${foo}`},
		common: &ParamExp{Param: lit("foo")},
	},
	{
		Strs: []string{`${foo}"bar"`},
		common: word(
			&ParamExp{Param: lit("foo")},
			dblQuoted(lit("bar")),
		),
	},
	{
		Strs: []string{`$a/b $a-b $a:b $a}b $a]b $a.b $a,b $a*b $a_b $a2b`},
		common: call(
			word(litParamExp("a"), lit("/b")),
			word(litParamExp("a"), lit("-b")),
			word(litParamExp("a"), lit(":b")),
			word(litParamExp("a"), lit("}b")),
			word(litParamExp("a"), lit("]b")),
			word(litParamExp("a"), lit(".b")),
			word(litParamExp("a"), lit(",b")),
			word(litParamExp("a"), lit("*b")),
			word(litParamExp("a_b")),
			word(litParamExp("a2b")),
		),
	},
	{
		Strs: []string{`$aàb $àb $,b`},
		common: call(
			word(litParamExp("a"), lit("àb")),
			word(lit("$"), lit("àb")),
			word(lit("$"), lit(",b")),
		),
	},
	{
		Strs: []string{"$foobar", "$foo\\\nbar"},
		common: call(
			word(litParamExp("foobar")),
		),
	},
	{
		Strs: []string{"$foo\\bar"},
		common: call(
			word(litParamExp("foo"), lit("\\bar")),
		),
	},
	{
		Strs: []string{`echo -e "$foo\nbar"`},
		common: call(
			litWord("echo"), litWord("-e"),
			word(dblQuoted(
				litParamExp("foo"), lit(`\nbar`),
			)),
		),
	},
	{
		Strs: []string{`${foo-bar}`},
		common: &ParamExp{
			Param: lit("foo"),
			Exp: &Expansion{
				Op:   SubstMinus,
				Word: litWord("bar"),
			},
		},
	},
	{
		Strs: []string{`${foo+}"bar"`},
		common: word(
			&ParamExp{
				Param: lit("foo"),
				Exp:   &Expansion{Op: SubstPlus},
			},
			dblQuoted(lit("bar")),
		),
	},
	{
		Strs: []string{`${foo:=<"bar"}`},
		common: &ParamExp{
			Param: lit("foo"),
			Exp: &Expansion{
				Op:   SubstColAssgn,
				Word: word(lit("<"), dblQuoted(lit("bar"))),
			},
		},
	},
	{
		Strs: []string{
			"${foo:=b${c}$(d)}",
			"${foo:=b${c}`d`}",
		},
		common: &ParamExp{
			Param: lit("foo"),
			Exp: &Expansion{
				Op: SubstColAssgn,
				Word: word(
					lit("b"),
					&ParamExp{Param: lit("c")},
					cmdSubst(litStmt("d")),
				),
			},
		},
	},
	{
		Strs: []string{`${foo?"${bar}"}`},
		common: &ParamExp{
			Param: lit("foo"),
			Exp: &Expansion{
				Op: SubstQuest,
				Word: word(dblQuoted(
					&ParamExp{Param: lit("bar")},
				)),
			},
		},
	},
	{
		Strs: []string{`${foo:?bar1 bar2}`},
		common: &ParamExp{
			Param: lit("foo"),
			Exp: &Expansion{
				Op:   SubstColQuest,
				Word: litWord("bar1 bar2"),
			},
		},
	},
	{
		Strs: []string{`${a:+b}${a:-b}${a=b}`},
		common: word(
			&ParamExp{
				Param: lit("a"),
				Exp: &Expansion{
					Op:   SubstColPlus,
					Word: litWord("b"),
				},
			},
			&ParamExp{
				Param: lit("a"),
				Exp: &Expansion{
					Op:   SubstColMinus,
					Word: litWord("b"),
				},
			},
			&ParamExp{
				Param: lit("a"),
				Exp: &Expansion{
					Op:   SubstAssgn,
					Word: litWord("b"),
				},
			},
		),
	},
	{
		Strs: []string{`${a:-'$x'}`},
		common: &ParamExp{
			Param: lit("a"),
			Exp: &Expansion{
				Op:   SubstColMinus,
				Word: word(sglQuoted("$x")),
			},
		},
	},
	{
		Strs: []string{`${a:-$x}`},
		common: &ParamExp{
			Param: lit("a"),
			Exp: &Expansion{
				Op:   SubstColMinus,
				Word: word(litParamExp("x")),
			},
		},
	},
	{
		Strs: []string{`${foo%bar}${foo%%bar*}`},
		common: word(
			&ParamExp{
				Param: lit("foo"),
				Exp: &Expansion{
					Op:   RemSmallSuffix,
					Word: litWord("bar"),
				},
			},
			&ParamExp{
				Param: lit("foo"),
				Exp: &Expansion{
					Op:   RemLargeSuffix,
					Word: litWord("bar*"),
				},
			},
		),
	},
	{
		Strs: []string{`${foo#bar}${foo##bar*}`},
		common: word(
			&ParamExp{
				Param: lit("foo"),
				Exp: &Expansion{
					Op:   RemSmallPrefix,
					Word: litWord("bar"),
				},
			},
			&ParamExp{
				Param: lit("foo"),
				Exp: &Expansion{
					Op:   RemLargePrefix,
					Word: litWord("bar*"),
				},
			},
		),
	},
	{
		Strs: []string{`${foo%?}`},
		common: &ParamExp{
			Param: lit("foo"),
			Exp: &Expansion{
				Op:   RemSmallSuffix,
				Word: litWord("?"),
			},
		},
	},
	{
		Strs: []string{
			`${foo[1]}`,
			`${foo[ 1 ]}`,
		},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Index: litWord("1"),
		},
	},
	{
		Strs: []string{`${foo[-1]}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Index: &UnaryArithm{
				Op: Minus,
				X:  litWord("1"),
			},
		},
	},
	{
		Strs: []string{`${foo[@]}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Index: litWord("@"),
		},
	},
	{
		Strs: []string{`${foo[*]-etc}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Index: litWord("*"),
			Exp: &Expansion{
				Op:   SubstMinus,
				Word: litWord("etc"),
			},
		},
	},
	{
		Strs: []string{`${foo[bar]}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Index: litWord("bar"),
		},
	},
	{
		Strs: []string{`${foo[$bar]}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Index: word(litParamExp("bar")),
		},
	},
	{
		Strs: []string{`${foo:1}`, `${foo: 1 }`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Slice: &Slice{Offset: litWord("1")},
		},
	},
	{
		Strs: []string{`${foo:1:2}`, `${foo: 1 : 2 }`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Slice: &Slice{
				Offset: litWord("1"),
				Length: litWord("2"),
			},
		},
	},
	{
		Strs: []string{`${foo:a:b}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Slice: &Slice{
				Offset: litWord("a"),
				Length: litWord("b"),
			},
		},
	},
	{
		Strs: []string{`${foo:1:-2}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Slice: &Slice{
				Offset: litWord("1"),
				Length: &UnaryArithm{Op: Minus, X: litWord("2")},
			},
		},
	},
	{
		Strs: []string{`${foo::+3}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Slice: &Slice{
				Length: &UnaryArithm{Op: Plus, X: litWord("3")},
			},
		},
	},
	{
		Strs: []string{`${foo: -1}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Slice: &Slice{
				Offset: &UnaryArithm{Op: Minus, X: litWord("1")},
			},
		},
	},
	{
		Strs: []string{`${foo: +2+3}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Slice: &Slice{
				Offset: &BinaryArithm{
					Op: Add,
					X:  &UnaryArithm{Op: Plus, X: litWord("2")},
					Y:  litWord("3"),
				},
			},
		},
	},
	{
		Strs: []string{`${foo:a?1:2:3}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Slice: &Slice{
				Offset: &BinaryArithm{
					Op: Quest,
					X:  litWord("a"),
					Y: &BinaryArithm{
						Op: Colon,
						X:  litWord("1"),
						Y:  litWord("2"),
					},
				},
				Length: litWord("3"),
			},
		},
	},
	{
		Strs: []string{`${foo/a/b}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl:  &Replace{Orig: litWord("a"), With: litWord("b")},
		},
	},
	{
		Strs: []string{"${foo/ /\t}"},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl:  &Replace{Orig: litWord(" "), With: litWord("\t")},
		},
	},
	{
		Strs: []string{`${foo/[/]-}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl:  &Replace{Orig: litWord("["), With: litWord("]-")},
		},
	},
	{
		Strs: []string{`${foo/bar/b/a/r}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl: &Replace{
				Orig: litWord("bar"),
				With: litWord("b/a/r"),
			},
		},
	},
	{
		Strs: []string{`${foo/$a/$b}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl: &Replace{
				Orig: word(litParamExp("a")),
				With: word(litParamExp("b")),
			},
		},
	},
	{
		Strs: []string{`${foo//b1/b2}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl: &Replace{
				All:  true,
				Orig: litWord("b1"),
				With: litWord("b2"),
			},
		},
	},
	{
		Strs: []string{`${foo///}`, `${foo//}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl:  &Replace{All: true},
		},
	},
	{
		Strs: []string{`${foo/-//}`},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl:  &Replace{Orig: litWord("-"), With: litWord("/")},
		},
	},
	{
		Strs: []string{
			`${foo//#/}`,
			`${foo//#}`,
		},
		bsmk: &ParamExp{
			Param: lit("foo"),
			Repl:  &Replace{All: true, Orig: litWord("#")},
		},
	},
	{
		Strs: []string{`${a^b} ${a^^b} ${a,b} ${a,,b}`},
		bash: call(
			word(&ParamExp{Param: lit("a"),
				Exp: &Expansion{
					Op:   UpperFirst,
					Word: litWord("b"),
				},
			}),
			word(&ParamExp{Param: lit("a"),
				Exp: &Expansion{
					Op:   UpperAll,
					Word: litWord("b"),
				},
			}),
			word(&ParamExp{Param: lit("a"),
				Exp: &Expansion{
					Op:   LowerFirst,
					Word: litWord("b"),
				},
			}),
			word(&ParamExp{Param: lit("a"),
				Exp: &Expansion{
					Op:   LowerAll,
					Word: litWord("b"),
				},
			}),
		),
	},
	{
		Strs: []string{`${a@E} ${b@a} ${@@Q}`},
		bsmk: call(
			word(&ParamExp{Param: lit("a"),
				Exp: &Expansion{
					Op:   OtherParamOps,
					Word: litWord("E"),
				},
			}),
			word(&ParamExp{Param: lit("b"),
				Exp: &Expansion{
					Op:   OtherParamOps,
					Word: litWord("a"),
				},
			}),
			word(&ParamExp{Param: lit("@"),
				Exp: &Expansion{
					Op:   OtherParamOps,
					Word: litWord("Q"),
				},
			}),
		),
	},
	{
		Strs: []string{`${#foo}`},
		common: &ParamExp{
			Length: true,
			Param:  lit("foo"),
		},
	},
	{
		Strs: []string{`${%foo}`},
		mksh: &ParamExp{
			Width: true,
			Param: lit("foo"),
		},
	},
	{
		Strs: []string{`${!foo} ${!bar[@]}`},
		bsmk: call(
			word(&ParamExp{
				Excl:  true,
				Param: lit("foo"),
			}),
			word(&ParamExp{
				Excl:  true,
				Param: lit("bar"),
				Index: litWord("@"),
			}),
		),
	},
	{
		Strs: []string{`${!foo*} ${!bar@}`},
		bsmk: call(
			word(&ParamExp{
				Excl:  true,
				Param: lit("foo"),
				Names: NamesPrefix,
			}),
			word(&ParamExp{
				Excl:  true,
				Param: lit("bar"),
				Names: NamesPrefixWords,
			}),
		),
	},
	{
		Strs: []string{`${#?}`},
		common: call(
			word(&ParamExp{Length: true, Param: lit("?")}),
		),
	},
	{
		Strs:   []string{`"${foo}"`},
		common: dblQuoted(&ParamExp{Param: lit("foo")}),
	},
	{
		Strs:   []string{`"(foo)"`},
		common: dblQuoted(lit("(foo)")),
	},
	{
		Strs: []string{`"${foo}>"`},
		common: dblQuoted(
			&ParamExp{Param: lit("foo")},
			lit(">"),
		),
	},
	{
		Strs:   []string{`"$(foo)"`, "\"`foo`\""},
		common: dblQuoted(cmdSubst(litStmt("foo"))),
	},
	{
		Strs: []string{
			`"$(foo bar)"`,
			`"$(foo  bar)"`,
			"\"`foo bar`\"",
			"\"`foo  bar`\"",
		},
		common: dblQuoted(cmdSubst(litStmt("foo", "bar"))),
	},
	{
		Strs:   []string{`'${foo}'`},
		common: sglQuoted("${foo}"),
	},
	{
		Strs:   []string{"$((1))"},
		common: arithmExp(litWord("1")),
	},
	{
		Strs: []string{"$((1 + 3))", "$((1+3))"},
		common: arithmExp(&BinaryArithm{
			Op: Add,
			X:  litWord("1"),
			Y:  litWord("3"),
		}),
	},
	{
		Strs: []string{`"$((foo))"`},
		common: dblQuoted(arithmExp(
			litWord("foo"),
		)),
	},
	{
		Strs: []string{`$((a)) b`},
		common: call(
			word(arithmExp(litWord("a"))),
			litWord("b"),
		),
	},
	{
		Strs: []string{`$((arr[0]++))`},
		bsmk: arithmExp(&UnaryArithm{
			Op: Inc, Post: true,
			X: word(&ParamExp{
				Short: true,
				Param: lit("arr"),
				Index: litWord("0"),
			}),
		}),
	},
	{
		Strs: []string{`$((++arr[0]))`},
		bsmk: arithmExp(&UnaryArithm{
			Op: Inc,
			X: word(&ParamExp{
				Short: true,
				Param: lit("arr"),
				Index: litWord("0"),
			}),
		}),
	},
	{
		Strs: []string{`$((${a:-1}))`},
		bsmk: arithmExp(word(&ParamExp{
			Param: lit("a"),
			Exp: &Expansion{
				Op:   SubstColMinus,
				Word: litWord("1"),
			},
		})),
	},
	{
		Strs: []string{"$((5 * 2 - 1))", "$((5*2-1))"},
		common: arithmExp(&BinaryArithm{
			Op: Sub,
			X: &BinaryArithm{
				Op: Mul,
				X:  litWord("5"),
				Y:  litWord("2"),
			},
			Y: litWord("1"),
		}),
	},
	{
		Strs: []string{"$((i | 13))"},
		common: arithmExp(&BinaryArithm{
			Op: Or,
			X:  litWord("i"),
			Y:  litWord("13"),
		}),
	},
	{
		Strs: []string{"$(((a) + ((b))))"},
		common: arithmExp(&BinaryArithm{
			Op: Add,
			X:  parenArit(litWord("a")),
			Y:  parenArit(parenArit(litWord("b"))),
		}),
	},
	{
		Strs: []string{
			"$((3 % 7))",
			"$((3\n% 7))",
			"$((3\\\n % 7))",
		},
		common: arithmExp(&BinaryArithm{
			Op: Rem,
			X:  litWord("3"),
			Y:  litWord("7"),
		}),
	},
	{
		Strs: []string{`"$((1 / 3))"`},
		common: dblQuoted(arithmExp(&BinaryArithm{
			Op: Quo,
			X:  litWord("1"),
			Y:  litWord("3"),
		})),
	},
	{
		Strs: []string{"$((2 ** 10))"},
		common: arithmExp(&BinaryArithm{
			Op: Pow,
			X:  litWord("2"),
			Y:  litWord("10"),
		}),
	},
	{
		Strs: []string{`$(((1) ^ 3))`},
		common: arithmExp(&BinaryArithm{
			Op: Xor,
			X:  parenArit(litWord("1")),
			Y:  litWord("3"),
		}),
	},
	{
		Strs: []string{`$((1 >> (3 << 2)))`},
		common: arithmExp(&BinaryArithm{
			Op: Shr,
			X:  litWord("1"),
			Y: parenArit(&BinaryArithm{
				Op: Shl,
				X:  litWord("3"),
				Y:  litWord("2"),
			}),
		}),
	},
	{
		Strs: []string{`$((-(1)))`},
		common: arithmExp(&UnaryArithm{
			Op: Minus,
			X:  parenArit(litWord("1")),
		}),
	},
	{
		Strs: []string{`$((i++))`},
		common: arithmExp(&UnaryArithm{
			Op:   Inc,
			Post: true,
			X:    litWord("i"),
		}),
	},
	{
		Strs:   []string{`$((--i))`},
		common: arithmExp(&UnaryArithm{Op: Dec, X: litWord("i")}),
	},
	{
		Strs:   []string{`$((!i))`},
		common: arithmExp(&UnaryArithm{Op: Not, X: litWord("i")}),
	},
	{
		Strs: []string{`$((-!+i))`},
		common: arithmExp(&UnaryArithm{
			Op: Minus,
			X: &UnaryArithm{
				Op: Not,
				X:  &UnaryArithm{Op: Plus, X: litWord("i")},
			},
		}),
	},
	{
		Strs: []string{`$((!!i))`},
		common: arithmExp(&UnaryArithm{
			Op: Not,
			X:  &UnaryArithm{Op: Not, X: litWord("i")},
		}),
	},
	{
		Strs: []string{`$((1 < 3))`},
		common: arithmExp(&BinaryArithm{
			Op: Lss,
			X:  litWord("1"),
			Y:  litWord("3"),
		}),
	},
	{
		Strs: []string{`$((i = 2))`, `$((i=2))`},
		common: arithmExp(&BinaryArithm{
			Op: Assgn,
			X:  litWord("i"),
			Y:  litWord("2"),
		}),
	},
	{
		Strs: []string{`((a[i] = 4))`, `((a[i]=4))`},
		bsmk: arithmCmd(&BinaryArithm{
			Op: Assgn,
			X: word(&ParamExp{
				Short: true,
				Param: lit("a"),
				Index: litWord("i"),
			}),
			Y: litWord("4"),
		}),
	},
	{
		Strs: []string{"$((a += 2, b -= 3))"},
		common: arithmExp(&BinaryArithm{
			Op: Comma,
			X: &BinaryArithm{
				Op: AddAssgn,
				X:  litWord("a"),
				Y:  litWord("2"),
			},
			Y: &BinaryArithm{
				Op: SubAssgn,
				X:  litWord("b"),
				Y:  litWord("3"),
			},
		}),
	},
	{
		Strs: []string{"$((a >>= 2, b <<= 3))"},
		common: arithmExp(&BinaryArithm{
			Op: Comma,
			X: &BinaryArithm{
				Op: ShrAssgn,
				X:  litWord("a"),
				Y:  litWord("2"),
			},
			Y: &BinaryArithm{
				Op: ShlAssgn,
				X:  litWord("b"),
				Y:  litWord("3"),
			},
		}),
	},
	{
		Strs: []string{"$((a == b && c > d))"},
		common: arithmExp(&BinaryArithm{
			Op: AndArit,
			X: &BinaryArithm{
				Op: Eql,
				X:  litWord("a"),
				Y:  litWord("b"),
			},
			Y: &BinaryArithm{
				Op: Gtr,
				X:  litWord("c"),
				Y:  litWord("d"),
			},
		}),
	},
	{
		Strs: []string{"$((a != b))"},
		common: arithmExp(&BinaryArithm{
			Op: Neq,
			X:  litWord("a"),
			Y:  litWord("b"),
		}),
	},
	{
		Strs: []string{"$((a &= b))"},
		common: arithmExp(&BinaryArithm{
			Op: AndAssgn,
			X:  litWord("a"),
			Y:  litWord("b"),
		}),
	},
	{
		Strs: []string{"$((a |= b))"},
		common: arithmExp(&BinaryArithm{
			Op: OrAssgn,
			X:  litWord("a"),
			Y:  litWord("b"),
		}),
	},
	{
		Strs: []string{"$((a %= b))"},
		common: arithmExp(&BinaryArithm{
			Op: RemAssgn,
			X:  litWord("a"),
			Y:  litWord("b"),
		}),
	},
	{
		Strs: []string{"$((a /= b))"},
		common: arithmExp(&BinaryArithm{
			Op: QuoAssgn,
			X:  litWord("a"),
			Y:  litWord("b"),
		}),
	},
	{
		Strs: []string{"$((a ^= b))"},
		common: arithmExp(&BinaryArithm{
			Op: XorAssgn,
			X:  litWord("a"),
			Y:  litWord("b"),
		}),
	},
	{
		Strs: []string{"$((i *= 3))"},
		common: arithmExp(&BinaryArithm{
			Op: MulAssgn,
			X:  litWord("i"),
			Y:  litWord("3"),
		}),
	},
	{
		Strs: []string{"$((2 >= 10))"},
		common: arithmExp(&BinaryArithm{
			Op: Geq,
			X:  litWord("2"),
			Y:  litWord("10"),
		}),
	},
	{
		Strs: []string{"$((foo ? b1 : b2))"},
		common: arithmExp(&BinaryArithm{
			Op: Quest,
			X:  litWord("foo"),
			Y: &BinaryArithm{
				Op: Colon,
				X:  litWord("b1"),
				Y:  litWord("b2"),
			},
		}),
	},
	{
		Strs: []string{`$((a <= (1 || 2)))`},
		common: arithmExp(&BinaryArithm{
			Op: Leq,
			X:  litWord("a"),
			Y: parenArit(&BinaryArithm{
				Op: OrArit,
				X:  litWord("1"),
				Y:  litWord("2"),
			}),
		}),
	},
	{
		Strs:   []string{"foo$", "foo$\n"},
		common: word(lit("foo"), lit("$")),
	},
	{
		Strs:  []string{`$''`},
		bsmk:  sglDQuoted(""),
		posix: word(lit("$"), sglQuoted("")),
	},
	{
		Strs:  []string{`$""`},
		bsmk:  dblDQuoted(),
		posix: word(lit("$"), dblQuoted()),
	},
	{
		Strs:  []string{`$'foo'`},
		bsmk:  sglDQuoted("foo"),
		posix: word(lit("$"), sglQuoted("foo")),
	},
	{
		Strs: []string{`$'f+oo${'`},
		bsmk: sglDQuoted("f+oo${"),
	},
	{
		Strs: []string{"$'foo bar`'"},
		bsmk: sglDQuoted("foo bar`"),
	},
	{
		Strs: []string{"$'a ${b} c'"},
		bsmk: sglDQuoted("a ${b} c"),
	},
	{
		Strs: []string{`$"a ${b} c"`},
		bsmk: dblDQuoted(
			lit("a "),
			&ParamExp{Param: lit("b")},
			lit(" c"),
		),
	},
	{
		Strs:   []string{`"a $b c"`},
		common: dblQuoted(lit("a "), litParamExp("b"), lit(" c")),
	},
	{
		Strs: []string{`$"a $b c"`},
		bsmk: dblDQuoted(
			lit("a "),
			litParamExp("b"),
			lit(" c"),
		),
	},
	{
		Strs: []string{"$'f\\'oo\n'"},
		bsmk: sglDQuoted("f\\'oo\n"),
	},
	{
		Strs:  []string{`$"foo"`},
		bsmk:  dblDQuoted(lit("foo")),
		posix: word(lit("$"), dblQuoted(lit("foo"))),
	},
	{
		Strs: []string{`$"foo$"`},
		bsmk: dblDQuoted(lit("foo"), lit("$")),
	},
	{
		Strs: []string{`$"foo bar"`},
		bsmk: dblDQuoted(lit("foo bar")),
	},
	{
		Strs: []string{`$'f\'oo'`},
		bsmk: sglDQuoted(`f\'oo`),
	},
	{
		Strs: []string{`$"f\"oo"`},
		bsmk: dblDQuoted(lit(`f\"oo`)),
	},
	{
		Strs:   []string{`"foo$"`},
		common: dblQuoted(lit("foo"), lit("$")),
	},
	{
		Strs:   []string{`"foo$$"`},
		common: dblQuoted(lit("foo"), litParamExp("$")),
	},
	{
		Strs: []string{"$(foo$)", "`foo$`"},
		common: cmdSubst(
			stmt(call(word(lit("foo"), lit("$")))),
		),
	},
	{
		Strs:   []string{"foo$bar"},
		common: word(lit("foo"), litParamExp("bar")),
	},
	{
		Strs:   []string{"foo$(bar)"},
		common: word(lit("foo"), cmdSubst(litStmt("bar"))),
	},
	{
		Strs:   []string{"foo${bar}"},
		common: word(lit("foo"), &ParamExp{Param: lit("bar")}),
	},
	{
		Strs:   []string{"'foo${bar'"},
		common: sglQuoted("foo${bar"),
	},
	{
		Strs: []string{"(foo)\nbar", "(foo); bar"},
		common: []Command{
			subshell(litStmt("foo")),
			litCall("bar"),
		},
	},
	{
		Strs: []string{"foo\n(bar)", "foo; (bar)"},
		common: []Command{
			litCall("foo"),
			subshell(litStmt("bar")),
		},
	},
	{
		Strs: []string{"foo\n(bar)", "foo; (bar)"},
		common: []Command{
			litCall("foo"),
			subshell(litStmt("bar")),
		},
	},
	{
		Strs: []string{
			"case $i in 1) foo ;; 2 | 3*) bar ;; esac",
			"case $i in 1) foo;; 2 | 3*) bar; esac",
			"case $i in (1) foo;; 2 | 3*) bar;; esac",
			"case $i\nin\n#etc\n1)\nfoo\n;;\n2 | 3*)\nbar\n;;\nesac",
		},
		common: &CaseClause{
			Word: word(litParamExp("i")),
			Items: []*CaseItem{
				{
					Op:       Break,
					Patterns: litWords("1"),
					StmtList: litStmts("foo"),
				},
				{
					Op:       Break,
					Patterns: litWords("2", "3*"),
					StmtList: litStmts("bar"),
				},
			},
		},
	},
	{
		Strs: []string{"case i in 1) a ;& 2) ;; esac"},
		bsmk: &CaseClause{
			Word: litWord("i"),
			Items: []*CaseItem{
				{
					Op:       Fallthrough,
					Patterns: litWords("1"),
					StmtList: litStmts("a"),
				},
				{Op: Break, Patterns: litWords("2")},
			},
		},
	},
	{
		Strs: []string{
			"case i in 1) a ;; esac",
			"case i { 1) a ;; }",
			"case i {\n1) a ;;\n}",
		},
		mksh: &CaseClause{
			Word: litWord("i"),
			Items: []*CaseItem{{
				Op:       Break,
				Patterns: litWords("1"),
				StmtList: litStmts("a"),
			}},
		},
	},
	{
		Strs: []string{"case i in 1) a ;;& 2) b ;; esac"},
		bash: &CaseClause{
			Word: litWord("i"),
			Items: []*CaseItem{
				{
					Op:       Resume,
					Patterns: litWords("1"),
					StmtList: litStmts("a"),
				},
				{
					Op:       Break,
					Patterns: litWords("2"),
					StmtList: litStmts("b"),
				},
			},
		},
	},
	{
		Strs: []string{"case i in 1) a ;| 2) b ;; esac"},
		mksh: &CaseClause{
			Word: litWord("i"),
			Items: []*CaseItem{
				{
					Op:       ResumeKorn,
					Patterns: litWords("1"),
					StmtList: litStmts("a"),
				},
				{
					Op:       Break,
					Patterns: litWords("2"),
					StmtList: litStmts("b"),
				},
			},
		},
	},
	{
		Strs: []string{"case $i in 1) cat <<EOF ;;\nfoo\nEOF\nesac"},
		common: &CaseClause{
			Word: word(litParamExp("i")),
			Items: []*CaseItem{{
				Op:       Break,
				Patterns: litWords("1"),
				StmtList: stmtList(&Stmt{
					Cmd: litCall("cat"),
					Redirs: []*Redirect{{
						Op:   Hdoc,
						Word: litWord("EOF"),
						Hdoc: litWord("foo\n"),
					}},
				}),
			}},
		},
	},
	{
		Strs: []string{"foo | while read a; do b; done"},
		common: &BinaryCmd{
			Op: Pipe,
			X:  litStmt("foo"),
			Y: stmt(&WhileClause{
				Cond: stmtList(
					litStmt("read", "a"),
				),
				Do: litStmts("b"),
			}),
		},
	},
	{
		Strs: []string{"while read l; do foo || bar; done"},
		common: &WhileClause{
			Cond: stmtList(litStmt("read", "l")),
			Do: stmts(&BinaryCmd{
				Op: OrStmt,
				X:  litStmt("foo"),
				Y:  litStmt("bar"),
			}),
		},
	},
	{
		Strs:   []string{"echo if while"},
		common: litCall("echo", "if", "while"),
	},
	{
		Strs:   []string{"${foo}if"},
		common: word(&ParamExp{Param: lit("foo")}, lit("if")),
	},
	{
		Strs:   []string{"$if'|'"},
		common: word(litParamExp("if"), sglQuoted("|")),
	},
	{
		Strs: []string{"if a; then b=; fi", "if a; then b=\nfi"},
		common: &IfClause{
			Cond: litStmts("a"),
			Then: stmtList(stmt(&CallExpr{
				Assigns: []*Assign{
					{Name: lit("b")},
				},
			})),
		},
	},
	{
		Strs: []string{"if a; then >f; fi", "if a; then >f\nfi"},
		common: &IfClause{
			Cond: litStmts("a"),
			Then: stmtList(&Stmt{
				Redirs: []*Redirect{
					{Op: RdrOut, Word: litWord("f")},
				},
			}),
		},
	},
	{
		Strs: []string{"if a; then (a); fi", "if a; then (a) fi"},
		common: &IfClause{
			Cond: litStmts("a"),
			Then: stmts(subshell(litStmt("a"))),
		},
	},
	{
		Strs: []string{"a=b\nc=d", "a=b; c=d"},
		common: []Command{
			&CallExpr{Assigns: []*Assign{
				{Name: lit("a"), Value: litWord("b")},
			}},
			&CallExpr{Assigns: []*Assign{
				{Name: lit("c"), Value: litWord("d")},
			}},
		},
	},
	{
		Strs: []string{"foo && write | read"},
		common: &BinaryCmd{
			Op: AndStmt,
			X:  litStmt("foo"),
			Y: stmt(&BinaryCmd{
				Op: Pipe,
				X:  litStmt("write"),
				Y:  litStmt("read"),
			}),
		},
	},
	{
		Strs: []string{"write | read && bar"},
		common: &BinaryCmd{
			Op: AndStmt,
			X: stmt(&BinaryCmd{
				Op: Pipe,
				X:  litStmt("write"),
				Y:  litStmt("read"),
			}),
			Y: litStmt("bar"),
		},
	},
	{
		Strs: []string{"foo >f | bar"},
		common: &BinaryCmd{
			Op: Pipe,
			X: &Stmt{
				Cmd: litCall("foo"),
				Redirs: []*Redirect{
					{Op: RdrOut, Word: litWord("f")},
				},
			},
			Y: litStmt("bar"),
		},
	},
	{
		Strs: []string{"(foo) >f | bar"},
		common: &BinaryCmd{
			Op: Pipe,
			X: &Stmt{
				Cmd: subshell(litStmt("foo")),
				Redirs: []*Redirect{
					{Op: RdrOut, Word: litWord("f")},
				},
			},
			Y: litStmt("bar"),
		},
	},
	{
		Strs: []string{"foo | >f"},
		common: &BinaryCmd{
			Op: Pipe,
			X:  litStmt("foo"),
			Y: &Stmt{Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("f")},
			}},
		},
	},
	{
		Strs:  []string{"[[ a ]]"},
		bsmk:  &TestClause{X: litWord("a")},
		posix: litStmt("[[", "a", "]]"),
	},
	{
		Strs: []string{"[[ a ]]\nb"},
		bsmk: stmts(
			&TestClause{X: litWord("a")},
			litCall("b"),
		),
	},
	{
		Strs: []string{"[[ a > b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: TsAfter,
			X:  litWord("a"),
			Y:  litWord("b"),
		}},
	},
	{
		Strs: []string{"[[ 1 -nt 2 ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: TsNewer,
			X:  litWord("1"),
			Y:  litWord("2"),
		}},
	},
	{
		Strs: []string{"[[ 1 -eq 2 ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: TsEql,
			X:  litWord("1"),
			Y:  litWord("2"),
		}},
	},
	{
		Strs: []string{"[[ -R a ]]"},
		bash: &TestClause{X: &UnaryTest{
			Op: TsRefVar,
			X:  litWord("a"),
		}},
	},
	{
		Strs: []string{"[[ a =~ b ]]", "[[ a =~ b ]];"},
		bash: &TestClause{X: &BinaryTest{
			Op: TsReMatch,
			X:  litWord("a"),
			Y:  litWord("b"),
		}},
	},
	{
		Strs: []string{`[[ a =~ " foo "$bar ]]`},
		bash: &TestClause{X: &BinaryTest{
			Op: TsReMatch,
			X:  litWord("a"),
			Y: word(
				dblQuoted(lit(" foo ")),
				litParamExp("bar"),
			),
		}},
	},
	{
		Strs: []string{`[[ a =~ [ab](c |d) ]]`},
		bash: &TestClause{X: &BinaryTest{
			Op: TsReMatch,
			X:  litWord("a"),
			Y:  litWord("[ab](c |d)"),
		}},
	},
	{
		Strs: []string{`[[ a =~ ( ]]) ]]`},
		bash: &TestClause{X: &BinaryTest{
			Op: TsReMatch,
			X:  litWord("a"),
			Y:  litWord("( ]])"),
		}},
	},
	{
		Strs: []string{`[[ a =~ b\ c ]]`},
		bash: &TestClause{X: &BinaryTest{
			Op: TsReMatch,
			X:  litWord("a"),
			Y:  litWord(`b\ c`),
		}},
	},
	{
		Strs: []string{`[[ a == -n ]]`},
		bsmk: &TestClause{X: &BinaryTest{
			Op: TsMatch,
			X:  litWord("a"),
			Y:  litWord("-n"),
		}},
	},
	{
		Strs: []string{`[[ a =~ -n ]]`},
		bash: &TestClause{X: &BinaryTest{
			Op: TsReMatch,
			X:  litWord("a"),
			Y:  litWord("-n"),
		}},
	},
	{
		Strs: []string{"[[ -n $a ]]"},
		bsmk: &TestClause{
			X: &UnaryTest{Op: TsNempStr, X: word(litParamExp("a"))},
		},
	},
	{
		Strs: []string{"[[ ! $a < 'b' ]]"},
		bsmk: &TestClause{X: &UnaryTest{
			Op: TsNot,
			X: &BinaryTest{
				Op: TsBefore,
				X:  word(litParamExp("a")),
				Y:  word(sglQuoted("b")),
			},
		}},
	},
	{
		Strs: []string{
			"[[ ! -e $a ]]",
			"[[ ! -a $a ]]",
		},
		bsmk: &TestClause{X: &UnaryTest{
			Op: TsNot,
			X:  &UnaryTest{Op: TsExists, X: word(litParamExp("a"))},
		}},
	},
	{
		Strs: []string{
			"[[ a && b ]]",
			"[[\na &&\nb ]]",
			"[[\n\na &&\n\nb ]]",
		},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  litWord("a"),
			Y:  litWord("b"),
		}},
	},
	{
		Strs: []string{"[[ (a && b) ]]"},
		bsmk: &TestClause{X: parenTest(&BinaryTest{
			Op: AndTest,
			X:  litWord("a"),
			Y:  litWord("b"),
		})},
	},
	{
		Strs: []string{"[[ (a && b) || -f c ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: OrTest,
			X: parenTest(&BinaryTest{
				Op: AndTest,
				X:  litWord("a"),
				Y:  litWord("b"),
			}),
			Y: &UnaryTest{Op: TsRegFile, X: litWord("c")},
		}},
	},
	{
		Strs: []string{
			"[[ -S a && -L b ]]",
			"[[ -S a && -h b ]]",
		},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsSocket, X: litWord("a")},
			Y:  &UnaryTest{Op: TsSmbLink, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -k a && -N b ]]"},
		bash: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsSticky, X: litWord("a")},
			Y:  &UnaryTest{Op: TsModif, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -G a && -O b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsGrpOwn, X: litWord("a")},
			Y:  &UnaryTest{Op: TsUsrOwn, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -d a && -c b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsDirect, X: litWord("a")},
			Y:  &UnaryTest{Op: TsCharSp, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -b a && -p b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsBlckSp, X: litWord("a")},
			Y:  &UnaryTest{Op: TsNmPipe, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -g a && -u b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsGIDSet, X: litWord("a")},
			Y:  &UnaryTest{Op: TsUIDSet, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -r a && -w b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsRead, X: litWord("a")},
			Y:  &UnaryTest{Op: TsWrite, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -x a && -s b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsExec, X: litWord("a")},
			Y:  &UnaryTest{Op: TsNoEmpty, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -t a && -z b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsFdTerm, X: litWord("a")},
			Y:  &UnaryTest{Op: TsEmpStr, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ -o a && -v b ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X:  &UnaryTest{Op: TsOptSet, X: litWord("a")},
			Y:  &UnaryTest{Op: TsVarSet, X: litWord("b")},
		}},
	},
	{
		Strs: []string{"[[ a -ot b && c -ef d ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X: &BinaryTest{
				Op: TsOlder,
				X:  litWord("a"),
				Y:  litWord("b"),
			},
			Y: &BinaryTest{
				Op: TsDevIno,
				X:  litWord("c"),
				Y:  litWord("d"),
			},
		}},
	},
	{
		Strs: []string{
			"[[ a == b && c != d ]]",
			"[[ a = b && c != d ]]",
		},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X: &BinaryTest{
				Op: TsMatch,
				X:  litWord("a"),
				Y:  litWord("b"),
			},
			Y: &BinaryTest{
				Op: TsNoMatch,
				X:  litWord("c"),
				Y:  litWord("d"),
			},
		}},
	},
	{
		Strs: []string{"[[ a -ne b && c -le d ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X: &BinaryTest{
				Op: TsNeq,
				X:  litWord("a"),
				Y:  litWord("b"),
			},
			Y: &BinaryTest{
				Op: TsLeq,
				X:  litWord("c"),
				Y:  litWord("d"),
			},
		}},
	},
	{
		Strs: []string{"[[ c -ge d ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: TsGeq,
			X:  litWord("c"),
			Y:  litWord("d"),
		}},
	},
	{
		Strs: []string{"[[ a -lt b && c -gt d ]]"},
		bsmk: &TestClause{X: &BinaryTest{
			Op: AndTest,
			X: &BinaryTest{
				Op: TsLss,
				X:  litWord("a"),
				Y:  litWord("b"),
			},
			Y: &BinaryTest{
				Op: TsGtr,
				X:  litWord("c"),
				Y:  litWord("d"),
			},
		}},
	},
	{
		Strs:   []string{"declare -f func"},
		common: litStmt("declare", "-f", "func"),
		bash: &DeclClause{
			Variant: lit("declare"),
			Opts:    litWords("-f"),
			Assigns: []*Assign{{
				Naked: true,
				Name:  lit("func"),
			}},
		},
	},
	{
		Strs: []string{"(local bar)"},
		bsmk: subshell(stmt(&DeclClause{
			Variant: lit("local"),
			Assigns: []*Assign{{
				Naked: true,
				Name:  lit("bar"),
			}},
		})),
		posix: subshell(litStmt("local", "bar")),
	},
	{
		Strs:  []string{"typeset"},
		bsmk:  &DeclClause{Variant: lit("typeset")},
		posix: litStmt("typeset"),
	},
	{
		Strs: []string{"export bar"},
		bsmk: &DeclClause{
			Variant: lit("export"),
			Assigns: []*Assign{{
				Naked: true,
				Name:  lit("bar"),
			}},
		},
		posix: litStmt("export", "bar"),
	},
	{
		Strs:  []string{"readonly -n"},
		bsmk:  &DeclClause{Variant: lit("readonly"), Opts: litWords("-n")},
		posix: litStmt("readonly", "-n"),
	},
	{
		Strs: []string{"nameref bar"},
		bsmk: &DeclClause{
			Variant: lit("nameref"),
			Assigns: []*Assign{{
				Naked: true,
				Name:  lit("bar"),
			}},
		},
		posix: litStmt("nameref", "bar"),
	},
	{
		Strs: []string{"declare -a +n -b$o foo=bar"},
		bash: &DeclClause{
			Variant: lit("declare"),
			Opts: []*Word{
				litWord("-a"),
				litWord("+n"),
				word(lit("-b"), litParamExp("o")),
			},
			Assigns: []*Assign{
				{Name: lit("foo"), Value: litWord("bar")},
			},
		},
	},
	{
		Strs: []string{
			"declare -a foo=(b1 $(b2))",
			"declare -a foo=(b1 `b2`)",
		},
		bash: &DeclClause{
			Variant: lit("declare"),
			Opts:    litWords("-a"),
			Assigns: []*Assign{{
				Name: lit("foo"),
				Array: arrValues(
					litWord("b1"),
					word(cmdSubst(litStmt("b2"))),
				),
			}},
		},
	},
	{
		Strs: []string{"local -a foo=(b1)"},
		bash: &DeclClause{
			Variant: lit("local"),
			Opts:    litWords("-a"),
			Assigns: []*Assign{{
				Name:  lit("foo"),
				Array: arrValues(litWord("b1")),
			}},
		},
	},
	{
		Strs: []string{"declare -A foo=([a]=b)"},
		bash: &DeclClause{
			Variant: lit("declare"),
			Opts:    litWords("-A"),
			Assigns: []*Assign{{
				Name: lit("foo"),
				Array: &ArrayExpr{Elems: []*ArrayElem{{
					Index: litWord("a"),
					Value: litWord("b"),
				}}},
			}},
		},
	},
	{
		Strs: []string{
			"declare foo[a]",
			"declare foo[a]=",
		},
		bash: &DeclClause{
			Variant: lit("declare"),
			Assigns: []*Assign{{
				Name:  lit("foo"),
				Index: litWord("a"),
			}},
		},
	},
	{
		Strs: []string{"declare foo[*]"},
		bash: &DeclClause{
			Variant: lit("declare"),
			Assigns: []*Assign{{
				Name:  lit("foo"),
				Index: litWord("*"),
			}},
		},
	},
	{
		Strs: []string{`declare foo["x y"]`},
		bash: &DeclClause{
			Variant: lit("declare"),
			Assigns: []*Assign{{
				Name:  lit("foo"),
				Index: word(dblQuoted(lit("x y"))),
			}},
		},
	},
	{
		Strs: []string{"foo=([)"},
		mksh: &CallExpr{Assigns: []*Assign{{
			Name:  lit("foo"),
			Array: arrValues(litWord("[")),
		}}},
	},
	{
		Strs: []string{
			"a && b=(c)\nd",
			"a && b=(c); d",
		},
		bsmk: stmts(
			&BinaryCmd{
				Op: AndStmt,
				X:  litStmt("a"),
				Y: stmt(&CallExpr{Assigns: []*Assign{{
					Name:  lit("b"),
					Array: arrValues(litWord("c")),
				}}}),
			},
			litCall("d"),
		),
	},
	{
		Strs: []string{"declare -f $func >/dev/null"},
		bash: &Stmt{
			Cmd: &DeclClause{
				Variant: lit("declare"),
				Opts:    litWords("-f"),
				Assigns: []*Assign{{
					Naked: true,
					Value: word(litParamExp("func")),
				}},
			},
			Redirs: []*Redirect{
				{Op: RdrOut, Word: litWord("/dev/null")},
			},
		},
	},
	{
		Strs: []string{"declare a\n{ x; }"},
		bash: stmts(
			&DeclClause{
				Variant: lit("declare"),
				Assigns: []*Assign{{
					Naked: true,
					Name:  lit("a"),
				}},
			},
			block(litStmt("x")),
		),
	},
	{
		Strs:   []string{"eval a=b foo"},
		common: litStmt("eval", "a=b", "foo"),
	},
	{
		Strs:  []string{"time", "time\n"},
		posix: litStmt("time"),
		bsmk:  &TimeClause{},
	},
	{
		Strs:  []string{"time -p"},
		posix: litStmt("time", "-p"),
		bsmk:  &TimeClause{PosixFormat: true},
	},
	{
		Strs:  []string{"time -a"},
		posix: litStmt("time", "-a"),
		bsmk:  &TimeClause{Stmt: litStmt("-a")},
	},
	{
		Strs:  []string{"time --"},
		posix: litStmt("time", "--"),
		bsmk:  &TimeClause{Stmt: litStmt("--")},
	},
	{
		Strs: []string{"time foo"},
		bsmk: &TimeClause{Stmt: litStmt("foo")},
	},
	{
		Strs: []string{"time { foo; }"},
		bsmk: &TimeClause{Stmt: stmt(block(litStmt("foo")))},
	},
	{
		Strs:   []string{"coproc foo bar"},
		common: litStmt("coproc", "foo", "bar"),
		bash:   &CoprocClause{Stmt: litStmt("foo", "bar")},
	},
	{
		Strs: []string{"coproc name { foo; }"},
		bash: &CoprocClause{
			Name: lit("name"),
			Stmt: stmt(block(litStmt("foo"))),
		},
	},
	{
		Strs: []string{"coproc foo", "coproc foo;"},
		bash: &CoprocClause{Stmt: litStmt("foo")},
	},
	{
		Strs: []string{"coproc { foo; }"},
		bash: &CoprocClause{
			Stmt: stmt(block(litStmt("foo"))),
		},
	},
	{
		Strs: []string{"coproc (foo)"},
		bash: &CoprocClause{
			Stmt: stmt(subshell(litStmt("foo"))),
		},
	},
	{
		Strs: []string{"coproc name foo | bar"},
		bash: &CoprocClause{
			Name: lit("name"),
			Stmt: stmt(&BinaryCmd{
				Op: Pipe,
				X:  litStmt("foo"),
				Y:  litStmt("bar"),
			}),
		},
	},
	{
		Strs: []string{"coproc $()", "coproc ``"},
		bash: &CoprocClause{Stmt: stmt(call(
			word(cmdSubst()),
		))},
	},
	{
		Strs: []string{`let i++`},
		bsmk: letClause(
			&UnaryArithm{Op: Inc, Post: true, X: litWord("i")},
		),
		posix: litStmt("let", "i++"),
	},
	{
		Strs: []string{`let a++ b++ c +d`},
		bsmk: letClause(
			&UnaryArithm{Op: Inc, Post: true, X: litWord("a")},
			&UnaryArithm{Op: Inc, Post: true, X: litWord("b")},
			litWord("c"),
			&UnaryArithm{Op: Plus, X: litWord("d")},
		),
	},
	{
		Strs: []string{`let ++i >/dev/null`},
		bsmk: &Stmt{
			Cmd:    letClause(&UnaryArithm{Op: Inc, X: litWord("i")}),
			Redirs: []*Redirect{{Op: RdrOut, Word: litWord("/dev/null")}},
		},
	},
	{
		Strs: []string{
			`let a=(1 + 2) b=3+4`,
			`let a=(1+2) b=3+4`,
		},
		bash: letClause(
			&BinaryArithm{
				Op: Assgn,
				X:  litWord("a"),
				Y: parenArit(&BinaryArithm{
					Op: Add,
					X:  litWord("1"),
					Y:  litWord("2"),
				}),
			},
			&BinaryArithm{
				Op: Assgn,
				X:  litWord("b"),
				Y: &BinaryArithm{
					Op: Add,
					X:  litWord("3"),
					Y:  litWord("4"),
				},
			},
		),
	},
	{
		Strs:   []string{"(foo-bar)"},
		common: subshell(litStmt("foo-bar")),
	},
	{
		Strs: []string{
			"let i++\nbar",
			"let i++ \nbar",
			"let i++; bar",
		},
		bsmk: []*Stmt{
			stmt(letClause(&UnaryArithm{
				Op:   Inc,
				Post: true,
				X:    litWord("i"),
			})),
			litStmt("bar"),
		},
	},
	{
		Strs: []string{
			"let i++\nfoo=(bar)",
			"let i++; foo=(bar)",
			"let i++; foo=(bar)\n",
		},
		bsmk: []*Stmt{
			stmt(letClause(&UnaryArithm{
				Op:   Inc,
				Post: true,
				X:    litWord("i"),
			})),
			stmt(&CallExpr{Assigns: []*Assign{{
				Name:  lit("foo"),
				Array: arrValues(litWord("bar")),
			}}}),
		},
	},
	{
		Strs: []string{
			"case a in b) let i++ ;; esac",
			"case a in b) let i++;; esac",
		},
		bsmk: &CaseClause{
			Word: word(lit("a")),
			Items: []*CaseItem{{
				Op:       Break,
				Patterns: litWords("b"),
				StmtList: stmtList(stmt(letClause(&UnaryArithm{
					Op:   Inc,
					Post: true,
					X:    litWord("i"),
				}))),
			}},
		},
	},
	{
		Strs: []string{"a+=1"},
		bsmk: &CallExpr{
			Assigns: []*Assign{{
				Append: true,
				Name:   lit("a"),
				Value:  litWord("1"),
			}},
		},
		posix: litStmt("a+=1"),
	},
	{
		Strs: []string{"b+=(2 3)"},
		bsmk: &CallExpr{Assigns: []*Assign{{
			Append: true,
			Name:   lit("b"),
			Array:  arrValues(litWords("2", "3")...),
		}}},
	},
	{
		Strs:  []string{"a[2]=b c[-3]= d[x]+=e"},
		posix: litStmt("a[2]=b", "c[-3]=", "d[x]+=e"),
		bsmk: &CallExpr{Assigns: []*Assign{
			{
				Name:  lit("a"),
				Index: litWord("2"),
				Value: litWord("b"),
			},
			{
				Name: lit("c"),
				Index: &UnaryArithm{
					Op: Minus,
					X:  litWord("3"),
				},
			},
			{
				Name:   lit("d"),
				Index:  litWord("x"),
				Append: true,
				Value:  litWord("e"),
			},
		}},
	},
	{
		Strs: []string{
			"b[i]+=2",
			"b[ i ]+=2",
		},
		bsmk: &CallExpr{Assigns: []*Assign{{
			Append: true,
			Name:   lit("b"),
			Index:  litWord("i"),
			Value:  litWord("2"),
		}}},
	},
	{
		Strs: []string{`echo ${a["x y"]}`},
		bash: call(litWord("echo"), word(&ParamExp{
			Param: lit("a"),
			Index: word(dblQuoted(lit("x y"))),
		})),
	},
	{
		Strs: []string{
			`a["x y"]=b`,
			`a[ "x y" ]=b`,
		},
		bash: &CallExpr{Assigns: []*Assign{{
			Name:  lit("a"),
			Index: word(dblQuoted(lit("x y"))),
			Value: litWord("b"),
		}}},
	},
	{
		Strs: []string{`((a["x y"] = b))`, `((a["x y"]=b))`},
		bsmk: arithmCmd(&BinaryArithm{
			Op: Assgn,
			X: word(&ParamExp{
				Short: true,
				Param: lit("a"),
				Index: word(dblQuoted(lit("x y"))),
			}),
			Y: litWord("b"),
		}),
	},
	{
		Strs: []string{
			`a=(["x y"]=b)`,
			`a=( [ "x y" ]=b)`,
		},
		bash: &CallExpr{Assigns: []*Assign{{
			Name: lit("a"),
			Array: &ArrayExpr{Elems: []*ArrayElem{{
				Index: word(dblQuoted(lit("x y"))),
				Value: litWord("b"),
			}}},
		}}},
	},
	{
		Strs:   []string{"a]b"},
		common: litStmt("a]b"),
	},
	{
		Strs:  []string{"echo a[b c[de]f"},
		posix: litStmt("echo", "a[b", "c[de]f"),
		bsmk: call(litWord("echo"),
			word(lit("a"), lit("[b")),
			word(lit("c"), lit("[de]f")),
		),
	},
	{
		Strs: []string{"<<EOF | b\nfoo\nEOF"},
		common: &BinaryCmd{
			Op: Pipe,
			X: &Stmt{Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("foo\n"),
			}}},
			Y: litStmt("b"),
		},
	},
	{
		Strs: []string{"<<EOF1 <<EOF2 | c && d\nEOF1\nEOF2"},
		common: &BinaryCmd{
			Op: AndStmt,
			X: stmt(&BinaryCmd{
				Op: Pipe,
				X: &Stmt{Redirs: []*Redirect{
					{Op: Hdoc, Word: litWord("EOF1")},
					{Op: Hdoc, Word: litWord("EOF2")},
				}},
				Y: litStmt("c"),
			}),
			Y: litStmt("d"),
		},
	},
	{
		Strs: []string{
			"<<EOF && { bar; }\nhdoc\nEOF",
			"<<EOF &&\nhdoc\nEOF\n{ bar; }",
		},
		common: &BinaryCmd{
			Op: AndStmt,
			X: &Stmt{Redirs: []*Redirect{{
				Op:   Hdoc,
				Word: litWord("EOF"),
				Hdoc: litWord("hdoc\n"),
			}}},
			Y: stmt(block(litStmt("bar"))),
		},
	},
	{
		Strs: []string{"foo() {\n\t<<EOF && { bar; }\nhdoc\nEOF\n}"},
		common: &FuncDecl{
			Name: lit("foo"),
			Body: stmt(block(stmt(&BinaryCmd{
				Op: AndStmt,
				X: &Stmt{Redirs: []*Redirect{{
					Op:   Hdoc,
					Word: litWord("EOF"),
					Hdoc: litWord("hdoc\n"),
				}}},
				Y: stmt(block(litStmt("bar"))),
			}))),
		},
	},
	{
		Strs: []string{`"a$("")"`, "\"a`\"\"`\""},
		common: dblQuoted(
			lit("a"),
			cmdSubst(stmt(call(
				word(dblQuoted()),
			))),
		),
	},
	{
		Strs: []string{"echo ?(b)*(c)+(d)@(e)!(f)"},
		bsmk: stmt(call(litWord("echo"), word(
			&ExtGlob{Op: GlobQuest, Pattern: lit("b")},
			&ExtGlob{Op: GlobStar, Pattern: lit("c")},
			&ExtGlob{Op: GlobPlus, Pattern: lit("d")},
			&ExtGlob{Op: GlobAt, Pattern: lit("e")},
			&ExtGlob{Op: GlobExcl, Pattern: lit("f")},
		))),
	},
	{
		Strs: []string{"echo foo@(b*(c|d))bar"},
		bsmk: stmt(call(litWord("echo"), word(
			lit("foo"),
			&ExtGlob{Op: GlobAt, Pattern: lit("b*(c|d)")},
			lit("bar"),
		))),
	},
	{
		Strs: []string{"echo $a@(b)$c?(d)$e*(f)$g+(h)$i!(j)$k"},
		bsmk: stmt(call(litWord("echo"), word(
			litParamExp("a"),
			&ExtGlob{Op: GlobAt, Pattern: lit("b")},
			litParamExp("c"),
			&ExtGlob{Op: GlobQuest, Pattern: lit("d")},
			litParamExp("e"),
			&ExtGlob{Op: GlobStar, Pattern: lit("f")},
			litParamExp("g"),
			&ExtGlob{Op: GlobPlus, Pattern: lit("h")},
			litParamExp("i"),
			&ExtGlob{Op: GlobExcl, Pattern: lit("j")},
			litParamExp("k"),
		))),
	},
}

// these don't have a canonical format with the same AST
var fileTestsNoPrint = []testCase{
	{
		Strs:  []string{`$[foo]`},
		posix: word(lit("$"), lit("[foo]")),
	},
	{
		Strs:  []string{`"$[foo]"`},
		posix: dblQuoted(lit("$"), lit("[foo]")),
	},
	{
		Strs: []string{`"$[1 + 3]"`},
		bash: dblQuoted(arithmExpBr(&BinaryArithm{
			Op: Add,
			X:  litWord("1"),
			Y:  litWord("3"),
		})),
	},
}

func fullProg(v interface{}) *File {
	f := &File{}
	switch x := v.(type) {
	case *File:
		return x
	case StmtList:
		f.Stmts = x.Stmts
		return f
	case []*Stmt:
		f.Stmts = x
		return f
	case *Stmt:
		f.Stmts = append(f.Stmts, x)
		return f
	case []Command:
		for _, cmd := range x {
			f.Stmts = append(f.Stmts, stmt(cmd))
		}
		return f
	case *Word:
		return fullProg(call(x))
	case WordPart:
		return fullProg(word(x))
	case Command:
		return fullProg(stmt(x))
	case nil:
	default:
		panic(reflect.TypeOf(v))
	}
	return nil
}

func clearPosRecurse(tb testing.TB, src string, v interface{}) {
	zeroPos := Pos{}
	checkSrc := func(pos Pos, strs ...string) {
		if src == "" {
			return
		}
		offs := pos.Offset()
		if offs > uint(len(src)) {
			tb.Fatalf("Pos %d in %T is out of bounds in %q",
				pos, v, string(src))
			return
		}
		if strs == nil {
			return
		}
		var gotErr string
		for i, want := range strs {
			got := src[offs:]
			if i == 0 {
				gotErr = got
			}
			got = strings.Replace(got, "\\\n", "", -1)
			if len(got) > len(want) {
				got = got[:len(want)]
			}
			if got == want {
				return
			}
		}
		tb.Fatalf("Expected one of %q at %d in %q, found %q",
			strs, pos, src, gotErr)
	}
	setPos := func(p *Pos, strs ...string) {
		checkSrc(*p, strs...)
		if *p == zeroPos {
			tb.Fatalf("Pos in %T is already %v", v, zeroPos)
		}
		*p = zeroPos
	}
	checkPos := func(n Node) {
		if n == nil {
			return
		}
		if n.Pos() != zeroPos {
			tb.Fatalf("Found unexpected Pos() in %T: want %d, got %d",
				n, zeroPos, n.Pos())
		}
		if n.Pos().After(n.End()) {
			tb.Fatalf("Found End() before Pos() in %T", n)
		}
	}
	recurse := func(v interface{}) {
		clearPosRecurse(tb, src, v)
		if n, ok := v.(Node); ok {
			checkPos(n)
		}
	}
	switch x := v.(type) {
	case *File:
		recurse(x.StmtList)
		checkPos(x)
	case []Comment:
		for i := range x {
			recurse(&x[i])
		}
	case *Comment:
		setPos(&x.Hash, "#"+x.Text)
	case StmtList:
		for _, s := range x.Stmts {
			recurse(s)
		}
		recurse(x.Last)
	case *Stmt:
		endOff := int(x.End().Offset())
		if endOff < len(src) {
			end := src[endOff]
			switch {
			case end == ' ', end == '\n', end == '\t', end == '\r':
				// ended by whitespace
			case regOps(rune(end)):
				// ended by end character
			case endOff > 0 && src[endOff-1] == ';':
				// ended by semicolon
			default:
				tb.Fatalf("Unexpected Stmt.End() %d %q in %q",
					endOff, end, string(src))
			}
		}
		recurse(x.Comments)
		setPos(&x.Position)
		if x.Semicolon.IsValid() {
			setPos(&x.Semicolon, ";")
		}
		if x.Cmd != nil {
			recurse(x.Cmd)
		}
		for _, r := range x.Redirs {
			setPos(&r.OpPos, r.Op.String())
			if r.N != nil {
				recurse(r.N)
			}
			recurse(r.Word)
			if r.Hdoc != nil {
				recurse(r.Hdoc)
			}
		}
	case []*Assign:
		for _, a := range x {
			if a.Name != nil {
				recurse(a.Name)
			}
			if a.Index != nil {
				recurse(a.Index)
			}
			if a.Value != nil {
				recurse(a.Value)
			}
			if a.Array != nil {
				recurse(a.Array)
			}
			checkPos(a)
		}
	case *CallExpr:
		recurse(x.Assigns)
		recurse(x.Args)
	case []*Word:
		for _, w := range x {
			recurse(w)
		}
	case *Word:
		recurse(x.Parts)
	case []WordPart:
		for _, wp := range x {
			recurse(wp)
		}
	case *Lit:
		pos, end := int(x.Pos().Offset()), int(x.End().Offset())
		want := pos + len(x.Value)
		val := x.Value
		posLine := x.Pos().Line()
		endLine := x.End().Line()
		switch {
		case src == "":
		case strings.Contains(src, "\\\n"):
		case !strings.Contains(x.Value, "\n") && posLine != endLine:
			tb.Fatalf("Lit without newlines has Pos/End lines %d and %d",
				posLine, endLine)
		case strings.Contains(src, "\\\\"):
			// removed quotes inside backquote cmd substs
			val = ""
		case end < len(src) && src[end] == '\n':
			// heredoc literals that end with the
			// stop word and a newline
		case end == len(src):
			// same as above, but with word and EOF
		case end != want:
			tb.Fatalf("Unexpected Lit.End() %d (wanted %d) in %q",
				end, want, string(src))
		}
		setPos(&x.ValuePos, val)
		setPos(&x.ValueEnd)
	case *Subshell:
		setPos(&x.Lparen, "(")
		setPos(&x.Rparen, ")")
		recurse(x.StmtList)
	case *Block:
		setPos(&x.Lbrace, "{")
		setPos(&x.Rbrace, "}")
		recurse(x.StmtList)
	case *IfClause:
		setPos(&x.IfPos, "if", "elif")
		setPos(&x.ThenPos, "then")
		if x.FiPos.IsValid() {
			setPos(&x.FiPos, "fi", "elif")
		}
		recurse(x.Cond)
		recurse(x.Then)
		if !x.Else.empty() {
			setPos(&x.ElsePos, "else", "elif")
			recurse(x.Else)
		}
	case *WhileClause:
		rsrv := "while"
		if x.Until {
			rsrv = "until"
		}
		setPos(&x.WhilePos, rsrv)
		setPos(&x.DoPos, "do")
		setPos(&x.DonePos, "done")
		recurse(x.Cond)
		recurse(x.Do)
	case *ForClause:
		if x.Select {
			setPos(&x.ForPos, "select")
		} else {
			setPos(&x.ForPos, "for")
		}
		setPos(&x.DoPos, "do")
		setPos(&x.DonePos, "done")
		recurse(x.Loop)
		recurse(x.Do)
	case *WordIter:
		recurse(x.Name)
		recurse(x.Items)
	case *CStyleLoop:
		setPos(&x.Lparen, "((")
		setPos(&x.Rparen, "))")
		if x.Init != nil {
			recurse(x.Init)
		}
		if x.Cond != nil {
			recurse(x.Cond)
		}
		if x.Post != nil {
			recurse(x.Post)
		}
	case *SglQuoted:
		valuePos := posAddCol(x.Left, 1)
		if x.Dollar {
			valuePos = posAddCol(valuePos, 1)
		}
		checkSrc(valuePos, x.Value)
		if x.Dollar {
			setPos(&x.Left, "$'")
		} else {
			setPos(&x.Left, "'")
		}
		setPos(&x.Right, "'")
	case *DblQuoted:
		checkSrc(posAddCol(x.End(), -1), `"`)
		if x.Dollar {
			setPos(&x.Position, `$"`)
		} else {
			setPos(&x.Position, `"`)
		}
		recurse(x.Parts)
	case *UnaryArithm:
		setPos(&x.OpPos, x.Op.String())
		recurse(x.X)
	case *UnaryTest:
		strs := []string{x.Op.String()}
		switch x.Op {
		case TsExists:
			strs = append(strs, "-a")
		case TsSmbLink:
			strs = append(strs, "-h")
		}
		setPos(&x.OpPos, strs...)
		recurse(x.X)
	case *BinaryCmd:
		setPos(&x.OpPos, x.Op.String())
		recurse(x.X)
		recurse(x.Y)
	case *BinaryArithm:
		setPos(&x.OpPos, x.Op.String())
		recurse(x.X)
		recurse(x.Y)
	case *BinaryTest:
		strs := []string{x.Op.String()}
		switch x.Op {
		case TsMatch:
			strs = append(strs, "=")
		}
		setPos(&x.OpPos, strs...)
		recurse(x.X)
		recurse(x.Y)
	case *ParenArithm:
		setPos(&x.Lparen, "(")
		setPos(&x.Rparen, ")")
		recurse(x.X)
	case *ParenTest:
		setPos(&x.Lparen, "(")
		setPos(&x.Rparen, ")")
		recurse(x.X)
	case *FuncDecl:
		if x.RsrvWord {
			setPos(&x.Position, "function")
		} else {
			setPos(&x.Position)
		}
		recurse(x.Name)
		recurse(x.Body)
	case *ParamExp:
		doll := "$"
		if x.nakedIndex() {
			doll = ""
		}
		setPos(&x.Dollar, doll)
		if !x.Short {
			setPos(&x.Rbrace, "}")
		} else if x.nakedIndex() {
			checkSrc(posAddCol(x.End(), -1), "]")
		}
		recurse(x.Param)
		if x.Index != nil {
			recurse(x.Index)
		}
		if x.Slice != nil {
			if x.Slice.Offset != nil {
				recurse(x.Slice.Offset)
			}
			if x.Slice.Length != nil {
				recurse(x.Slice.Length)
			}
		}
		if x.Repl != nil {
			if x.Repl.Orig != nil {
				recurse(x.Repl.Orig)
			}
			if x.Repl.With != nil {
				recurse(x.Repl.With)
			}
		}
		if x.Exp != nil && x.Exp.Word != nil {
			recurse(x.Exp.Word)
		}
	case *ArithmExp:
		if x.Bracket {
			// deprecated $(( form
			setPos(&x.Left, "$[")
			setPos(&x.Right, "]")
		} else {
			setPos(&x.Left, "$((")
			setPos(&x.Right, "))")
		}
		recurse(x.X)
	case *ArithmCmd:
		setPos(&x.Left, "((")
		setPos(&x.Right, "))")
		recurse(x.X)
	case *CmdSubst:
		switch {
		case x.TempFile:
			setPos(&x.Left, "${ ", "${\t", "${\n")
			setPos(&x.Right, "}")
		case x.ReplyVar:
			setPos(&x.Left, "${|")
			setPos(&x.Right, "}")
		default:
			setPos(&x.Left, "$(", "`", "\\`")
			setPos(&x.Right, ")", "`", "\\`")
		}
		recurse(x.StmtList)
	case *CaseClause:
		setPos(&x.Case, "case")
		setPos(&x.Esac, "esac", "}")
		recurse(x.Word)
		for _, ci := range x.Items {
			recurse(ci)
		}
	case *CaseItem:
		setPos(&x.OpPos, x.Op.String(), "esac")
		recurse(x.Patterns)
		recurse(x.StmtList)
	case *TestClause:
		setPos(&x.Left, "[[")
		setPos(&x.Right, "]]")
		recurse(x.X)
	case *DeclClause:
		recurse(x.Variant)
		recurse(x.Opts)
		recurse(x.Assigns)
	case *TimeClause:
		setPos(&x.Time, "time")
		if x.Stmt != nil {
			recurse(x.Stmt)
		}
	case *CoprocClause:
		setPos(&x.Coproc, "coproc")
		if x.Name != nil {
			recurse(x.Name)
		}
		recurse(x.Stmt)
	case *LetClause:
		setPos(&x.Let, "let")
		for _, expr := range x.Exprs {
			recurse(expr)
		}
	case *ArrayExpr:
		setPos(&x.Lparen, "(")
		setPos(&x.Rparen, ")")
		for _, elem := range x.Elems {
			recurse(elem)
		}
	case *ArrayElem:
		if x.Index != nil {
			recurse(x.Index)
		}
		recurse(x.Value)
	case *ExtGlob:
		setPos(&x.OpPos, x.Op.String())
		checkSrc(posAddCol(x.End(), -1), ")")
		recurse(x.Pattern)
	case *ProcSubst:
		setPos(&x.OpPos, x.Op.String())
		setPos(&x.Rparen, ")")
		recurse(x.StmtList)
	default:
		panic(reflect.TypeOf(v))
	}
}
