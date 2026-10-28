package narra

// The clone functions copy only the syntax of a function. Fields the checker
// fills in are left empty, so each generic instance is checked from scratch.

func cloneFunc(fd *FuncDecl) *FuncDecl {
	c := &FuncDecl{Pos: fd.Pos, Name: fd.Name, Exported: fd.Exported, Result: fd.Result, Effects: fd.Effects}
	for _, p := range fd.Params {
		c.Params = append(c.Params, &Param{Pos: p.Pos, Name: p.Name, Mutable: p.Mutable, Escaping: p.Escaping, Type: p.Type})
	}
	c.Body = cloneBlock(fd.Body)
	return c
}

func cloneBlock(b *Block) *Block {
	if b == nil {
		return nil
	}
	c := &Block{Pos: b.Pos}
	for _, s := range b.Stmts {
		c.Stmts = append(c.Stmts, cloneStmt(s))
	}
	return c
}

func cloneStmt(s Stmt) Stmt {
	switch s := s.(type) {
	case *Block:
		return cloneBlock(s)
	case *VarStmt:
		return &VarStmt{Pos: s.Pos, Mutable: s.Mutable, Name: s.Name, Pattern: clonePattern(s.Pattern), Type: s.Type, Value: cloneExpr(s.Value)}
	case *AssignStmt:
		return &AssignStmt{Pos: s.Pos, Target: cloneExpr(s.Target), Value: cloneExpr(s.Value)}
	case *ExprStmt:
		return &ExprStmt{Pos: s.Pos, Expr: cloneExpr(s.Expr)}
	case *ReturnStmt:
		return &ReturnStmt{Pos: s.Pos, Value: cloneExpr(s.Value)}
	case *IfStmt:
		return &IfStmt{Pos: s.Pos, Cond: cloneExpr(s.Cond), Then: cloneBlock(s.Then), Else: cloneStmt(s.Else)}
	case *WhileStmt:
		return &WhileStmt{Pos: s.Pos, Cond: cloneExpr(s.Cond), Body: cloneBlock(s.Body)}
	case *ForStmt:
		return &ForStmt{Pos: s.Pos, Name: s.Name, Iter: cloneExpr(s.Iter), Body: cloneBlock(s.Body)}
	case *BreakStmt:
		return &BreakStmt{Pos: s.Pos}
	case *ContinueStmt:
		return &ContinueStmt{Pos: s.Pos}
	}
	return nil
}

func clonePattern(p *Pattern) *Pattern {
	if p == nil {
		return nil
	}
	c := &Pattern{Pos: p.Pos, Kind: p.Kind, Name: p.Name, Enum: p.Enum, Lit: cloneExpr(p.Lit)}
	if p.Args != nil {
		c.Args = []*Pattern{}
		for _, a := range p.Args {
			c.Args = append(c.Args, clonePattern(a))
		}
	}
	for _, f := range p.Fields {
		c.Fields = append(c.Fields, &FieldPattern{Pos: f.Pos, Name: f.Name, Pattern: clonePattern(f.Pattern)})
	}
	return c
}

func cloneExpr(e Expr) Expr {
	if e == nil {
		return nil
	}
	switch e := e.(type) {
	case *IntLit:
		return &IntLit{exprBase: exprBase{P: e.P}, Text: e.Text}
	case *FloatLit:
		return &FloatLit{exprBase: exprBase{P: e.P}, Text: e.Text}
	case *StringLit:
		return &StringLit{exprBase: exprBase{P: e.P}, Value: e.Value}
	case *CharLit:
		return &CharLit{exprBase: exprBase{P: e.P}, Value: e.Value}
	case *BoolLit:
		return &BoolLit{exprBase: exprBase{P: e.P}, Value: e.Value}
	case *TemplateLit:
		c := &TemplateLit{exprBase: exprBase{P: e.P}}
		for _, p := range e.Parts {
			c.Parts = append(c.Parts, cloneExpr(p))
		}
		return c
	case *Ident:
		return &Ident{exprBase: exprBase{P: e.P}, Name: e.Name}
	case *FieldExpr:
		return &FieldExpr{exprBase: exprBase{P: e.P}, Target: cloneExpr(e.Target), Name: e.Name}
	case *IndexExpr:
		return &IndexExpr{exprBase: exprBase{P: e.P}, Target: cloneExpr(e.Target), Index: cloneExpr(e.Index)}
	case *CallExpr:
		c := &CallExpr{exprBase: exprBase{P: e.P}, Callee: cloneExpr(e.Callee)}
		for _, a := range e.Args {
			c.Args = append(c.Args, &Arg{Mutable: a.Mutable, Value: cloneExpr(a.Value)})
		}
		return c
	case *UnaryExpr:
		return &UnaryExpr{exprBase: exprBase{P: e.P}, Op: e.Op, Operand: cloneExpr(e.Operand)}
	case *BinaryExpr:
		return &BinaryExpr{exprBase: exprBase{P: e.P}, Op: e.Op, Left: cloneExpr(e.Left), Right: cloneExpr(e.Right)}
	case *StructLit:
		c := &StructLit{exprBase: exprBase{P: e.P}, TypeName: e.TypeName, Spread: cloneExpr(e.Spread)}
		for _, f := range e.Fields {
			c.Fields = append(c.Fields, &FieldInit{Pos: f.Pos, Name: f.Name, Value: cloneExpr(f.Value)})
		}
		return c
	case *ArrayLit:
		c := &ArrayLit{exprBase: exprBase{P: e.P}, TupleCall: e.TupleCall}
		for _, el := range e.Elems {
			c.Elems = append(c.Elems, cloneExpr(el))
		}
		return c
	case *MatchExpr:
		c := &MatchExpr{exprBase: exprBase{P: e.P}, Subject: cloneExpr(e.Subject)}
		for _, a := range e.Arms {
			c.Arms = append(c.Arms, &MatchArm{Pos: a.Pos, Pattern: clonePattern(a.Pattern), Guard: cloneExpr(a.Guard), Body: cloneExpr(a.Body), Block: cloneBlock(a.Block)})
		}
		return c
	case *IsExpr:
		return &IsExpr{exprBase: exprBase{P: e.P}, Subject: cloneExpr(e.Subject), Pattern: clonePattern(e.Pattern)}
	case *ArrowExpr:
		return cloneArrow(e)
	}
	panic("clone: unknown expression")
}
