/*
 * gomacro - A Go interpreter with Lisp-like macros
 *
 * Copyright (C) 2017 Massimiliano Ghilardi
 *
 *     This program is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU Lesser General Public License as published
 *     by the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *
 *     This program is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU Lesser General Public License for more details.
 *
 *     You should have received a copy of the GNU Lesser General Public License
 *     along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 *
 * quasiquote.go
 *
 *  Created on Jun 09, 2017
 *      Author Massimiliano Ghilardi
 */

package fast

import (
	"go/ast"
	r "reflect"

	. "github.com/cosmos72/gomacro/ast2"
	. "github.com/cosmos72/gomacro/base"
	mt "github.com/cosmos72/gomacro/token"
)

func (c *Comp) QuasiquoteUnary(unary *ast.UnaryExpr) *Expr {
	block := unary.X.(*ast.FuncLit).Body

	// we invoke SimplifyNodeForQuote() at the end, not at the beginning.
	// reason: to support quasiquote{unquote_splice ...}
	toUnwrap := block != SimplifyNodeForQuote(block, true)

	in := ToAst(block)
	fun := c.Quasiquote(in).AsX1()

	return exprX1(c.TypeOfInterface(), func(env *Env) r.Value {
		node := AnyToAstWithNode(fun(env), "Quasiquote").Node()
		return r.ValueOf(SimplifyNodeForQuote(node, toUnwrap))
	})
}

// Quasiquote expands and compiles ~quasiquote, if Ast starts with it
func (c *Comp) Quasiquote(in Ast) *Expr {
	switch form := in.(type) {
	case UnaryExpr:
		if form.Op() == mt.QUASIQUOTE {
			body := form.X.X.(*ast.FuncLit).Body
			return c.quasiquote(ToAst(body))
		}
	}
	return c.Compile(in)
}

func (c *Comp) quasiquoteSlice(in Ast) *Expr {
	debug := c.Options&OptDebugMacroExpand != 0
	switch form := in.(type) {
	case UnaryExpr:
		switch op := form.Op(); op {
		case mt.UNQUOTE:
			node := SimplifyNodeForQuote(form.X.X.(*ast.FuncLit).Body, true)
			if debug {
				c.Debugf("Quasiquote slice expanding %s: %v", mt.String(op), node)
			}
			return c.CompileNode(node)
		case mt.UNQUOTE_SPLICE:
			body := form.X.X.(*ast.FuncLit).Body
			if debug {
				c.Debugf("Quasiquote slice expanding %s: %v", mt.String(op), body)
			}
			return c.CompileNode(body)
		}
	}
	return c.quasiquote(in)
}

// quasiquote expands and compiles the contents of a ~quasiquote
func (c *Comp) quasiquote(in Ast) *Expr {
	debug := c.Options&OptDebugMacroExpand != 0
	if debug {
		c.Debugf("Quasiquote expanding %s: %v", mt.String(mt.QUASIQUOTE), in.Interface())
	}
	switch in := in.(type) {
	case AstWithSlice:
		n := in.Size()
		funs := make([]func(*Env) r.Value, n)
		for i := 0; i < n; i++ {
			funs[i] = c.quasiquoteSlice(in.Get(i)).AsX1()
		}
		form := in.New().(AstWithSlice)

		return exprX1(c.TypeOf(form), func(env *Env) r.Value {
			out := form.New().(AstWithSlice)
			for _, fun := range funs {
				out.Append(AnyToAst(fun(env).Interface(), "Quasiquote"))
			}
			return r.ValueOf(out)
		})
	case UnaryExpr:
		switch op := in.Op(); op {
		case mt.UNQUOTE:
			node := SimplifyNodeForQuote(in.X.X.(*ast.FuncLit).Body, true)
			if debug {
				c.Debugf("Quasiquote expanding %s: %v", mt.String(op), node)
			}
			return c.CompileNode(node)
		case mt.UNQUOTE_SPLICE:
			c.Pos = in.X.Pos()
			c.Errorf("Quasiquote: cannot %s in single-node context: %v", mt.String(in.Op()), in.X)
			return nil
		}
	}

	// Ast can still be a tree: just not a resizeable one, so support ~unquote but not ~unquote_splice
	if in, ok := in.(AstWithNode); !ok {
		x := in.Interface()
		c.Errorf("Quasiquote: unsupported node type, expecting AstWithNode or AstWithSlice: %v <%v>", x, r.TypeOf(x))
		return nil
	} else {
		form := in.New().(AstWithNode) // clone input argument, do NOT retain it
		n := in.Size()
		if n == 0 {
			return exprX1(c.TypeOfInterface(), func(env *Env) r.Value {
				return r.ValueOf(form.New())
			})
		}
		funs := make([]func(*Env) r.Value, n)
		for i := 0; i < n; i++ {
			funs[i] = c.quasiquote(in.Get(i)).AsX1()
		}

		return exprX1(c.TypeOfInterface(), func(env *Env) r.Value {
			out := form.New().(AstWithNode)
			for i, fun := range funs {
				out.Set(i, AnyToAst(fun(env).Interface(), "Quasiquote"))
			}
			return r.ValueOf(out)
		})
	}
}
