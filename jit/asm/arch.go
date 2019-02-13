/*
 * gomacro - A Go interpreter with Lisp-like macros
 *
 * Copyright (C) 2018 Massimiliano Ghilardi
 *
 *     This Source Code Form is subject to the terms of the Mozilla Public
 *     License, v. 2.0. If a copy of the MPL was not distributed with this
 *     file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 *
 * arch.go
 *
 *  Created on Feb 13, 2019
 *      Author Massimiliano Ghilardi
 */

package asm

type ArchId uint8

const (
	NOARCH ArchId = iota
	ARM64
	AMD64
)

type Arch interface {
	Id() ArchId
	Name() string
	RLo() RegId
	RHi() RegId
	RegIdAlwaysLive() RegIds
	RegIdString(id RegId) string // RegId -> string
	RegIdValid(id RegId) bool
	RegString(r Reg) string // Reg -> string
	RegValid(r Reg) bool

	Init(asm *Asm, saveStart, saveEnd SaveSlot) *Asm
	Prologue(asm *Asm) *Asm
	Epilogue(asm *Asm) *Asm

	Op0(asm *Asm, op Op0) *Asm
	Op1(asm *Asm, op Op1, dst Arg) *Asm
	Op2(asm *Asm, op Op2, src Arg, dst Arg) *Asm
	Op3(asm *Asm, op Op3, a Arg, b Arg, dst Arg) *Asm
	Op4(asm *Asm, op Op4, a Arg, b Arg, c Arg, dst Arg) *Asm

	Zero(asm *Asm, dst Arg) *Asm
	Mov(asm *Asm, src Arg, dst Arg) *Asm
	Load(asm *Asm, src Mem, dst Reg) *Asm
	Store(asm *Asm, src Reg, dst Mem) *Asm
	Cast(asm *Asm, src Arg, dst Arg) *Asm
}

var Archs = []Arch{} // {Arm64{}, Amd64{}}