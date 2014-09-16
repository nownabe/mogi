package mogi

import (
	"reflect"
	// "database/sql"
	"database/sql/driver"

	// "github.com/davecgh/go-spew/spew"
	"github.com/youtube/vitess/go/vt/sqlparser"
)

type cond interface {
	matches(in input) bool
}

type condchain []cond

func (chain condchain) matches(in input) bool {
	for _, c := range chain {
		if !c.matches(in) {
			return false
		}
	}
	return true
}

type selectCond struct {
	cols []string
}

func (sc selectCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Select)
	if !ok {
		return false
	}

	// zero parameters means anything
	if len(sc.cols) == 0 {
		return true
	}

	return reflect.DeepEqual(sc.cols, in.cols())
}

type fromCond struct {
	tables []string
}

func (fc fromCond) matches(in input) bool {
	var inTables []string
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		for _, tex := range x.From {
			extractTableNames(&inTables, tex)
		}
	}
	return reflect.DeepEqual(fc.tables, inTables)
}

type tableCond struct {
	table string
}

func (tc tableCond) matches(in input) bool {
	switch x := in.statement.(type) {
	case *sqlparser.Insert:
		return tc.table == string(x.Table.Name)
	case *sqlparser.Update:
		return tc.table == string(x.Table.Name)
	case *sqlparser.Delete:
		return tc.table == string(x.Table.Name)
	}
	return false
}

type whereCond struct {
	col string
	v   interface{}
}

func newWhereCond(col string, v interface{}) whereCond {
	return whereCond{
		col: col,
		v:   unify(v),
	}
}

func (wc whereCond) matches(in input) bool {
	vals := in.where()
	v, ok := vals[wc.col]
	if !ok {
		return false
	}
	return reflect.DeepEqual(wc.v, v)
}

type argsCond struct {
	args []driver.Value
}

func (ac argsCond) matches(in input) bool {
	given := unifyArray(ac.args)
	return reflect.DeepEqual(given, in.args)
}

type insertCond struct {
	cols []string
}

func (ic insertCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Insert)
	if !ok {
		return false
	}

	// zero parameters means anything
	if len(ic.cols) == 0 {
		return true
	}

	return reflect.DeepEqual(ic.cols, in.cols())
}

type valueCond struct {
	row int
	col string
	v   interface{}
}

func newValueCond(row int, col string, v interface{}) valueCond {
	return valueCond{
		row: row,
		col: col,
		v:   unify(v),
	}
}

func (vc valueCond) matches(in input) bool {
	switch in.statement.(type) {
	case *sqlparser.Insert:
		values := in.rows()
		if vc.row > len(values)-1 {
			return false
		}
		v, ok := values[vc.row][vc.col]
		if !ok {
			return false
		}
		return reflect.DeepEqual(vc.v, v)
	case *sqlparser.Update:
		values := in.values()
		v, ok := values[vc.col]
		if !ok {
			return false
		}
		return reflect.DeepEqual(vc.v, v)
	}
	return false
}

type updateCond struct {
	cols []string
}

func (uc updateCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Update)
	if !ok {
		return false
	}

	// zero parameters means anything
	if len(uc.cols) == 0 {
		return true
	}

	return reflect.DeepEqual(uc.cols, in.cols())
}
