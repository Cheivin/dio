package wrapper

func Set(field string, val interface{}) *Update {
	return U().Set(field, val)
}

func SetIf(test bool, field string, val interface{}) *Update {
	return U().SetIf(test, field, val)
}

func SetExprIf(test bool, field string, Expr string, args ...interface{}) *Update {
	return U().SetExprIf(test, field, Expr, args...)
}

func SetExpr(field string, Expr string, args ...interface{}) *Update {
	return U().SetExpr(field, Expr, args...)
}
