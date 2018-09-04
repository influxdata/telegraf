package reql_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
    r "gopkg.in/gorethink/gorethink.v3"
	"gopkg.in/gorethink/gorethink.v3/internal/compare"
)

// ${description}
func Test${module_name}Suite(t *testing.T) {
	suite.Run(t, new(${module_name}Suite ))
}

type ${module_name}Suite struct {
	suite.Suite

	session *r.Session
}

func (suite *${module_name}Suite) SetupTest() {
	suite.T().Log("Setting up ${module_name}Suite")
	// Use imports to prevent errors
	_ = time.Time{}
    _ = compare.AnythingIsFine

	session, err := r.Connect(r.ConnectOpts{
		Address: url,
	})
	suite.Require().NoError(err, "Error returned when connecting to server")
	suite.session = session

	r.DBDrop("test").Exec(suite.session)
	err = r.DBCreate("test").Exec(suite.session)
	suite.Require().NoError(err)
	err = r.DB("test").Wait().Exec(suite.session)
	suite.Require().NoError(err)

	%for var_name in table_var_names:
	r.DB("test").TableDrop("${var_name}").Exec(suite.session)
	err = r.DB("test").TableCreate("${var_name}").Exec(suite.session)
	suite.Require().NoError(err)
	err = r.DB("test").Table("${var_name}").Wait().Exec(suite.session)
	suite.Require().NoError(err)
	%endfor
}

func (suite *${module_name}Suite) TearDownSuite() {
	suite.T().Log("Tearing down ${module_name}Suite")

	if suite.session != nil {
		r.DB("rethinkdb").Table("_debug_scratch").Delete().Exec(suite.session)
		%for var_name in table_var_names:
		 r.DB("test").TableDrop("${var_name}").Exec(suite.session)
		%endfor
		r.DBDrop("test").Exec(suite.session)

		suite.session.Close()
	}
}

<%rendered_vars = set() %>\
func (suite *${module_name}Suite) TestCases() {
	suite.T().Log("Running ${module_name}Suite: ${description}")

	%for var_name in table_var_names:
	${var_name} := r.DB("test").Table("${var_name}")
	_ = ${var_name} // Prevent any noused variable errors
	%endfor

<%rendered_something = False %>\
	%for item in defs_and_test:
	%if type(item) == GoDef:
<%rendered_something = True %>
	// ${item.testfile} line #${item.line_num}
	// ${item.line.original.replace('\n', '')}
	suite.T().Log("Possibly executing: ${item.line.go.replace('\\', '\\\\').replace('"', "'")}")

	%if item.varname in rendered_vars:
	%if item.run_if_query:
	${item.varname} = maybeRun(${item.value}, suite.session, r.RunOpts{
		%if item.runopts:
		%for key, val in sorted(item.runopts.items()):
		${key}: ${val},
		%endfor
		%endif
	});
	%else:
	${item.varname} = ${item.value}
	%endif
	%elif item.run_if_query:
	${item.varname} := maybeRun(${item.value}, suite.session, r.RunOpts{
		%if item.runopts:
		%for key, val in sorted(item.runopts.items()):
		${key}: ${val},
		%endfor
		%endif
	});
	_ = ${item.varname} // Prevent any noused variable errors
<%rendered_vars.add(item.varname)%>\
	%else:
	${item.varname} := ${item.value}
	_ = ${item.varname} // Prevent any noused variable errors
<%rendered_vars.add(item.varname)%>\
	%endif

	%elif type(item) == GoQuery:
<%rendered_something = True %>
	{
		// ${item.testfile} line #${item.line_num}
		/* ${item.expected_line.original} */
		var expected_ ${item.expected_type} = ${item.expected_line.go}
		/* ${item.line.original} */

		suite.T().Log("About to run line #${item.line_num}: ${item.line.go.replace('"', "'").replace('\\', '\\\\').replace('\n', '\\n')}")

		%if item.line.go.startswith('fetch(') and item.line.go.endswith(')'):
		fetchAndAssert(suite.Suite, expected_, ${item.line.go[6:-1]})
	   %elif item.is_value:
		actual := ${item.line.go}

		compare.Assert(suite.T(), expected_, actual)
		%else:
		runAndAssert(suite.Suite, expected_, ${item.line.go}, suite.session, r.RunOpts{
			%if item.runopts:
			%for key, val in sorted(item.runopts.items()):
			${key}: ${val},
			%endfor
			%endif
		})
		%endif
		suite.T().Log("Finished running line #${item.line_num}")
	}
	%endif
	%endfor
	%if not rendered_something:
<% raise EmptyTemplate() %>\
	%endif
}
