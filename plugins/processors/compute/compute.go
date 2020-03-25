package compute

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"go/ast"
	"go/parser"
	"go/token"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  ## Strategy to handle missing variables in cases where a formula refers to
  ## a non-existing field. Possible values are:
  ##		 ignore  - ignore formula for metric and do not update field
  ##     const   - target field will be replaced by the "constant" defined below
  ##     default - target field will be set to "default" defined below
  missing = "ignore"

  ## Constant to be used in the "const" strategy for missing fields
  # constant = 0

  ## Default value to be used in the "default" strategy for missing fields
  # default = 0

  ## Table of computations
  [processors.compute.fields]
    value = "(a + 3) / 4.3"
    x_sqr = "pow(a, 2)"
    x_abs = "abs(value)"
`

const (
	_ = iota
	ErrorBasicLiteral
	ErrorBinaryExpression
	ErrorFieldUnknown
	ErrorFunctionArguments
	ErrorFunctionTypeUnknown
	ErrorFunctionUnknown
	ErrorNodeTypeUnknown
	ErrorUnaryExpression
)

type ComputeError struct {
	Code byte
	Text string
}

func (e *ComputeError) Error() string {
	return e.Text
}

type Compute struct {
	Missing  string            `toml:"missing"`
	Constant interface{}       `toml:"constant"`
	Default  interface{}       `toml:"default"`
	Fields   map[string]string `toml:"fields"`

	trees map[string]ast.Expr
}

func (c *Compute) SampleConfig() string {
	return sampleConfig
}

func (c *Compute) Description() string {
	return "Compute values for a metric using the given formula(s)"
}

func (c *Compute) Init() error {
	// Check the given parameters
	switch c.Missing {
	case "ignore":
		c.Constant = nil
		c.Default = nil
	case "const":
		c.Default = nil

		if c.Constant == nil {
			c.Constant = int64(0)
		}

		switch c.Constant.(type) {
		case int, int64, float64:
		default:
			return fmt.Errorf("wrong datatype for 'constant', has to be int or float")
		}
	case "default":
		c.Constant = nil

		if c.Default == nil {
			c.Default = int64(0)
		}
		switch c.Default.(type) {
		case int:
			c.Default = int64(c.Default.(int))
		case int64, float64:
		default:
			return fmt.Errorf("wrong datatype for 'default', has to be int64 or float64")
		}
	default:
		return fmt.Errorf("unknown missing-value strategy '%s'", c.Missing)
	}

	// Parse all defined formulas into abstract-syntax-trees
	c.trees = make(map[string]ast.Expr, len(c.Fields))
	for name, expr := range c.Fields {
		log.Printf("D! [processors.compute] parsing formula for field \"%s\"", name)
		tree, err := parser.ParseExpr(expr)
		if err != nil {
			return fmt.Errorf("parsing of formula for field '%v' failed: %v", name, err)
		}
		c.trees[name] = tree
	}

	return nil
}

func (c *Compute) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		for name, tree := range c.trees {
			log.Printf("D! [processors.compute] processing field '%s' for metric '%s'", name, metric.Name())
			result, err := descend(tree, metric.Fields(), c.Default)
			if err == nil {
				metric.AddField(name, result)
			} else if err.(*ComputeError).Code == ErrorFieldUnknown {
				if c.Missing == "const" {
					metric.AddField(name, c.Constant)
				} else if c.Missing == "default" {
					metric.AddField(name, c.Default)
				}
			}
		}
	}
	return metrics
}

func handle_literal(kind token.Token, value string) (r interface{}, err error) {
	switch kind {
	case token.FLOAT:
		r, err = strconv.ParseFloat(value, 64)
	case token.INT:
		r, err = strconv.ParseInt(value, 10, 64)
	default:
		r, err = math.NaN(), &ComputeError{ErrorBasicLiteral, fmt.Sprintf("unknown basic literal '%v' with value '%v'", kind, value)}
	}

	return r, err
}

func binary_expr_int(operation token.Token, x, y int64) (r int64, err error) {
	switch operation {
	case token.ADD:
		r, err = x+y, nil
	case token.SUB:
		r, err = x-y, nil
	case token.MUL:
		r, err = x*y, nil
	case token.QUO:
		r, err = x/y, nil
	case token.REM:
		r, err = x%y, nil
	default:
		r, err = 0, &ComputeError{ErrorBinaryExpression, fmt.Sprintf("unknown binary integer operation '%v'", operation)}
	}

	return r, err
}

func binary_expr_float(operation token.Token, x, y float64) (r float64, err error) {
	switch operation {
	case token.ADD:
		r, err = x+y, nil
	case token.SUB:
		r, err = x-y, nil
	case token.MUL:
		r, err = x*y, nil
	case token.QUO:
		r, err = x/y, nil
	default:
		r, err = math.NaN(), &ComputeError{ErrorBinaryExpression, fmt.Sprintf("unknown binary float operation '%v'", operation)}
	}

	return r, err
}

func handle_binary_expression(operation token.Token, x, y interface{}) (r interface{}, err error) {
	_, x_int := x.(int64)
	_, y_int := y.(int64)
	if x_int && y_int {
		r, err = binary_expr_int(operation, x.(int64), y.(int64))
	} else {
		// Handle mixed float int arguments
		var fx, fy float64
		if x_int {
			fx = float64(x.(int64))
		} else {
			fx = x.(float64)
		}
		if y_int {
			fy = float64(y.(int64))
		} else {
			fy = y.(float64)
		}
		r, err = binary_expr_float(operation, fx, fy)
	}

	return r, err
}

func unary_expr_int(operation token.Token, x int64) (r int64, err error) {
	switch operation {
	case token.ADD:
		r, err = x, nil
	case token.SUB:
		r, err = -x, nil
	default:
		r, err = 0, &ComputeError{ErrorUnaryExpression, fmt.Sprintf("unknown unary int operation '%v'", operation)}
	}

	return r, err
}

func unary_expr_float(operation token.Token, x float64) (r float64, err error) {
	switch operation {
	case token.ADD:
		r, err = x, nil
	case token.SUB:
		r, err = -x, nil
	default:
		r, err = math.NaN(), &ComputeError{ErrorUnaryExpression, fmt.Sprintf("unknown unary float operation '%v'", operation)}
	}

	return r, err
}

func handle_unary_expression(operation token.Token, x interface{}) (r interface{}, err error) {
	_, x_int := x.(int64)
	if x_int {
		r, err = unary_expr_int(operation, x.(int64))
	} else {
		r, err = unary_expr_float(operation, x.(float64))
	}

	return r, err
}

func handle_function(fun string, args []interface{}) (r interface{}, err error) {
	// NOTE: When adding a function, please also add a documentation in the
	//       "Supported operations" section of the README!
	switch strings.ToLower(fun) {
	case "abs":
		if len(args) != 1 {
			return math.NaN(), &ComputeError{ErrorFunctionArguments, fmt.Sprintf("invalid number of arguments (%v) for function '%v'", len(args), fun)}
		}
		// Handle int/float arguments
		if _, x_int := args[0].(int64); x_int {
			x := args[0].(int64)
			if x >= 0 {
				r, err = x, nil
			} else {
				r, err = -x, nil
			}
		} else {
			r, err = math.Abs(args[0].(float64)), nil
		}
	case "pow":
		if len(args) != 2 {
			return math.NaN(), &ComputeError{ErrorFunctionArguments, fmt.Sprintf("invalid number of arguments (%v) for function '%v'", len(args), fun)}
		}
		var x, y float64

		// Convert arguments to float
		if _, x_int := args[0].(int64); x_int {
			x = float64(args[0].(int64))
		} else {
			x = args[0].(float64)
		}

		if _, y_int := args[1].(int64); y_int {
			y = float64(args[1].(int64))
		} else {
			y = args[1].(float64)
		}
		r, err = math.Pow(x, y), nil
	default:
		r, err = math.NaN(), &ComputeError{ErrorFunctionUnknown, fmt.Sprintf("unknown function '%v'", fun)}
	}

	return r, err
}

func descend(n ast.Node, fields map[string]interface{}, default_value interface{}) (r interface{}, err error) {
	switch d := n.(type) {
	case *ast.BasicLit:
		return handle_literal(d.Kind, d.Value)
	case *ast.BinaryExpr:
		x, err := descend(d.X, fields, default_value)
		if err != nil {
			return math.NaN(), err
		}
		y, err := descend(d.Y, fields, default_value)
		if err != nil {
			return math.NaN(), err
		}
		return handle_binary_expression(d.Op, x, y)
	case *ast.UnaryExpr:
		x, err := descend(d.X, fields, default_value)
		if err != nil {
			return math.NaN(), err
		}
		return handle_unary_expression(d.Op, x)
	case *ast.Ident:
		if _, ok := fields[d.Name]; !ok {
			if default_value != nil {
				return default_value, nil
			}
			return math.NaN(), &ComputeError{ErrorFieldUnknown, fmt.Sprintf("unknown field '%v'", d.Name)}
		}
		return fields[d.Name], nil
	case *ast.ParenExpr:
		r, err = descend(d.X, fields, default_value)
	case *ast.CallExpr:
		if fun, ok := d.Fun.(*ast.Ident); ok {
			// Handle the arguments
			args := make([]interface{}, len(d.Args))
			for i, a := range d.Args {
				x, err := descend(a, fields, default_value)
				if err != nil {
					return math.NaN(), err
				}
				args[i] = x
			}
			return handle_function(fun.Name, args)
		} else {
			return math.NaN(), &ComputeError{ErrorFunctionTypeUnknown, fmt.Sprintf("unknown function type '%v'", d.Fun)}
		}
	default:
		return math.NaN(), &ComputeError{ErrorNodeTypeUnknown, fmt.Sprintf("unknown node type '%v'", d)}
	}
	return r, err
}

func init() {
	processors.Add("compute", func() telegraf.Processor { return &Compute{} })
}
