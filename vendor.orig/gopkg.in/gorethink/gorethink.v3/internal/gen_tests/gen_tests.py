#!/usr/bin/env python3
# -*- coding: utf-8 -*-
'''Finds yaml tests, converts them to Go tests.'''
from __future__ import print_function

import sys
import os
import os.path
import re
import time
import ast
from re import sub, match, split, DOTALL
import argparse
import codecs
import logging
import process_polyglot
import parse_polyglot
from process_polyglot import Unhandled, Skip, FatalSkip, SkippedTest
from mako.template import Template

try:
    from cStringIO import StringIO
except ImportError:
    from io import StringIO
from collections import namedtuple

parse_polyglot.printDebug = False
logger = logging.getLogger("convert_tests")
r = None


TEST_EXCLUSIONS = [
    # python only tests
    # 'regression/1133',
    # 'regression/767',
    # 'regression/1005',
    'regression/',
    'limits',  # pending fix in issue #4965
    # double run
    'changefeeds/squash',
    'arity',
    '.rb.yaml',
]

GO_KEYWORDS = [
    'break', 'case', 'chan', 'const', 'continue', 'default', 'defer', 'else',
    'fallthrough', 'for', 'func', 'go', 'goto', 'if', 'import', 'interface',
    'map', 'package', 'range', 'return', 'select', 'struct', 'switch', 'type',
    'var'
]

def main():
    logging.basicConfig(format="[%(name)s] %(message)s", level=logging.INFO)
    start = time.clock()
    args = parse_args()
    if args.debug:
        logger.setLevel(logging.DEBUG)
        logging.getLogger('process_polyglot').setLevel(logging.DEBUG)
    elif args.info:
        logger.setLevel(logging.INFO)
        logging.getLogger('process_polyglot').setLevel(logging.INFO)
    else:
        logger.root.setLevel(logging.WARNING)
    if args.e:
        evaluate_snippet(args.e)
        exit(0)
    global r
    r = import_python_driver(args.python_driver_dir)
    renderer = Renderer(
        invoking_filenames=[
            __file__,
            process_polyglot.__file__,
        ])
    for testfile in process_polyglot.all_yaml_tests(
            args.test_dir,
            TEST_EXCLUSIONS):
        logger.info("Working on %s", testfile)
        TestFile(
            test_dir=args.test_dir,
            filename=testfile,
            test_output_dir=args.test_output_dir,
            renderer=renderer,
        ).load().render()
    logger.info("Finished in %s seconds", time.clock() - start)


def parse_args():
    '''Parse command line arguments'''
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--test-dir",
        help="Directory where yaml tests are",
        default=""
    )
    parser.add_argument(
        "--test-output-dir",
        help="Directory to render tests to",
        default=".",
    )
    parser.add_argument(
        "--python-driver-dir",
        help="Where the built python driver is located",
        default=""
    )
    parser.add_argument(
        "--test-file",
        help="Only convert the specified yaml file",
    )
    parser.add_argument(
        '--debug',
        help="Print debug output",
        dest='debug',
        action='store_true')
    parser.set_defaults(debug=False)
    parser.add_argument(
        '--info',
        help="Print info level output",
        dest='info',
        action='store_true')
    parser.set_defaults(info=False)
    parser.add_argument(
        '-e',
        help="Convert an inline python reql to go reql snippet",
    )
    return parser.parse_args()


def import_python_driver(py_driver_dir):
    '''Imports the test driver header'''
    stashed_path = sys.path
    sys.path.insert(0, os.path.realpath(py_driver_dir))
    import rethinkdb as r
    sys.path = stashed_path
    return r

GoQuery = namedtuple(
    'GoQuery',
    ('is_value',
     'line',
     'expected_type',
     'expected_line',
     'testfile',
     'line_num',
     'runopts')
)
GoDef = namedtuple(
    'GoDef',
    ('line',
     'varname',
     'vartype',
     'value',
     'run_if_query',
     'testfile',
     'line_num',
     'runopts')
)
Version = namedtuple("Version", "original go")

GO_DECL = re.compile(r'var (?P<var>\w+) (?P<type>.+) = (?P<value>.*)')


def evaluate_snippet(snippet):
    '''Just converts a single expression snippet into java'''
    try:
        parsed = ast.parse(snippet, mode='eval').body
    except Exception as e:
        return print("Error:", e)
    try:
        print(ReQLVisitor(smart_bracket=True).convert(parsed))
    except Exception as e:
        return print("Error:", e)


class TestFile(object):
    '''Represents a single test file'''

    def __init__(self, test_dir, filename, test_output_dir, renderer):
        self.filename = filename
        self.full_path = os.path.join(test_dir, filename)
        self.module_name = filename.split('.')[0].replace('/', '_')
        self.test_output_dir = test_output_dir
        self.reql_vars = {'r'}
        self.renderer = renderer

    def load(self):
        '''Load the test file, yaml parse it, extract file-level metadata'''
        with open(self.full_path, encoding='utf-8') as f:
            parsed_yaml = parse_polyglot.parseYAML(f)
        self.description = parsed_yaml.get('desc', 'No description')
        self.table_var_names = self.get_varnames(parsed_yaml)
        self.reql_vars.update(self.table_var_names)
        self.raw_test_data = parsed_yaml['tests']
        self.test_generator = process_polyglot.tests_and_defs(
            self.filename,
            self.raw_test_data,
            context=process_polyglot.create_context(r, self.table_var_names),
            custom_field='go',
        )
        return self

    def get_varnames(self, yaml_file):
        '''Extract table variable names from yaml variable
        They can be specified just space separated, or comma separated'''
        raw_var_names = yaml_file.get('table_variable_name', '')
        if not raw_var_names:
            return set()
        return set(re.split(r'[, ]+', raw_var_names))

    def render(self):
        '''Renders the converted tests to a runnable test file'''
        defs_and_test = ast_to_go(self.test_generator, self.reql_vars)
        self.renderer.source_files = [self.full_path]
        self.renderer.render(
            'template.go',
            output_dir=self.test_output_dir,
            output_name='reql_' + self.module_name + '_test.go',
            dependencies=[self.full_path],
            defs_and_test=defs_and_test,
            table_var_names=list(sorted(self.table_var_names)),
            module_name=camel(self.module_name),
            GoQuery=GoQuery,
            GoDef=GoDef,
            description=self.description,
        )


def py_to_go_type(py_type):
    '''Converts python types to their Go equivalents'''
    if py_type is None:
        return None
    elif isinstance(py_type, str):
        # This can be called on something already converted
        return py_type
    elif py_type.__name__ == 'function':
        return 'func()'
    elif (py_type.__module__ == 'datetime' and
          py_type.__name__ == 'datetime'):
        return 'time.Time'
    elif py_type.__module__ == 'builtins':
        if py_type.__name__.endswith('Error'):
            return 'error'

        return {
            bool: 'bool',
            bytes: '[]byte',
            int: 'int',
            float: 'float64',
            str: 'string',
            dict: 'map[interface{}]interface{}',
            list: '[]interface{}',
            object: 'map[interface{}]interface{}',
            type(None): 'interface{}',
        }[py_type]
    elif py_type.__module__ == 'rethinkdb.ast':
        return "r.Term"
        # Anomalous non-rule based capitalization in the python driver
        # return {
        # }.get(py_type.__name__, py_type.__name__)
    elif py_type.__module__ == 'rethinkdb.errors':
        return py_type.__name__
    elif py_type.__module__ == '?test?':
        return {
            'int_cmp': 'int',
            'float_cmp': 'float64',
            'err_regex': 'Err',
            'partial': 'compare.Expected',
            'bag': 'compare.Expected',
            'uuid': 'compare.Regex',  # clashes with ast.Uuid
        }.get(py_type.__name__, camel(py_type.__name__))
    elif py_type.__module__ == 'rethinkdb.query':
        # All of the constants like minval maxval etc are defined in
        # query.py, but no type name is provided to `type`, so we have
        # to pull it out of a class variable
        return camel(py_type.st)
    else:
        raise Unhandled(
            "Don't know how to convert python type {}.{} to Go"
            .format(py_type.__module__, py_type.__name__))


def def_to_go(item, reql_vars):
    if is_reql(item.term.type):
        reql_vars.add(item.varname)
    try:
        if is_reql(item.term.type):
            visitor = ReQLVisitor
        else:
            visitor = GoVisitor
        go_line = visitor(reql_vars,
                            type_=item.term.type,
                            is_def=True,
                            ).convert(item.term.ast)
    except Skip as skip:
        return SkippedTest(line=item.term.line, reason=str(skip))
    go_decl = GO_DECL.match(go_line).groupdict()
    return GoDef(
        line=Version(
            original=item.term.line,
            go=go_line,
        ),
        varname=go_decl['var'],
        vartype=go_decl['type'],
        value=go_decl['value'],
        run_if_query=item.run_if_query,
        testfile=item.testfile,
        line_num=item.line_num,
        runopts=convert_runopts(reql_vars, go_decl['type'], item.runopts)
    )


def query_to_go(item, reql_vars):
    if item.runopts is not None:
        converted_runopts = convert_runopts(
            reql_vars, item.query.type, item.runopts)
    else:
        converted_runopts = item.runopts
    if converted_runopts is None:
        converted_runopts = {}
    converted_runopts['GeometryFormat'] = '\"raw\"'
    if 'GroupFormat' not in converted_runopts:
        converted_runopts['GroupFormat'] = '\"map\"'
    try:
        is_value = False
        if not item.query.type.__module__.startswith("rethinkdb.") or (not item.query.type.__module__.startswith("rethinkdb.") and item.query.type.__name__.endswith("Error")):
            is_value = True
        go_line = ReQLVisitor(
            reql_vars, type_=item.query.type).convert(item.query.ast)
        if is_reql(item.expected.type):
            visitor = ReQLVisitor
        else:
            visitor = GoVisitor
        go_expected_line = visitor(
            reql_vars, type_=item.expected.type)\
            .convert(item.expected.ast)
    except Skip as skip:
        return SkippedTest(line=item.query.line, reason=str(skip))
    return GoQuery(
        is_value=is_value,
        line=Version(
            original=item.query.line,
            go=go_line,
        ),
        expected_type=py_to_go_type(item.expected.type),
        expected_line=Version(
            original=item.expected.line,
            go=go_expected_line,
        ),
        testfile=item.testfile,
        line_num=item.line_num,
        runopts=converted_runopts,
    )


def ast_to_go(sequence, reql_vars):
    '''Converts the the parsed test data to go source lines using the
    visitor classes'''
    reql_vars = set(reql_vars)
    for item in sequence:
        if type(item) == process_polyglot.Def:
            yield def_to_go(item, reql_vars)
        elif type(item) == process_polyglot.CustomDef:
            yield GoDef(line=Version(item.line, item.line),
                          testfile=item.testfile,
                          line_num=item.line_num)
        elif type(item) == process_polyglot.Query:
            yield query_to_go(item, reql_vars)
        elif type(item) == SkippedTest:
            yield item
        else:
            assert False, "shouldn't happen, item was {}".format(item)


def is_reql(t):
    '''Determines if a type is a reql term'''
    # Other options for module: builtins, ?test?, datetime
    if not hasattr(t, '__module__'):
        return True

    return t.__module__ == 'rethinkdb.ast'


def escape_string(s, out):
    out.write('"')
    for codepoint in s:
        rpr = repr(codepoint)[1:-1]
        if rpr.startswith('\\x'):
            rpr = '\\u00' + rpr[2:]
        elif rpr == '"':
            rpr = r'\"'
        out.write(rpr)
    out.write('"')


def attr_matches(path, node):
    '''Helper function. Several places need to know if they are an
    attribute of some root object'''
    root, name = path.split('.')
    ret = is_name(root, node.value) and node.attr == name
    return ret


def is_name(name, node):
    '''Determine if the current attribute node is a Name with the
    given name'''
    return type(node) == ast.Name and node.id == name


def convert_runopts(reql_vars, type_, runopts):
    if runopts is None:
        return None
    return {
        camel(key): GoVisitor(
            reql_vars, type_=type_).convert(val)
        for key, val in runopts.items()
    }


class GoVisitor(ast.NodeVisitor):
    '''Converts python ast nodes into a Go string'''

    def __init__(self,
                 reql_vars=frozenset("r"),
                 out=None,
                 type_=None,
                 is_def=False,
                 smart_bracket=False,
    ):
        self.out = StringIO() if out is None else out
        self.reql_vars = reql_vars
        self.type = py_to_go_type(type_)
        self._type = type_
        self.is_def = is_def
        self.smart_bracket = smart_bracket
        super(GoVisitor, self).__init__()
        self.write = self.out.write

    def skip(self, message, *args, **kwargs):
        cls = Skip
        is_fatal = kwargs.pop('fatal', False)
        if self.is_def or is_fatal:
            cls = FatalSkip
        raise cls(message, *args, **kwargs)

    def convert(self, node):
        '''Convert a text line to another text line'''
        self.visit(node)
        return self.out.getvalue()

    def join(self, sep, items):
        first = True
        for item in items:
            if first:
                first = False
            else:
                self.write(sep)
            self.visit(item)

    def to_str(self, s):
        escape_string(s, self.out)

    def cast_null(self, arg):
        if (type(arg) == ast.Name and arg.id == 'null') or \
           (type(arg) == ast.NameConstant and arg.value == None):
            self.write("nil")
        else:
            self.visit(arg)

    def to_args(self, args, func='', optargs=[]):
        optargs_first = False

        self.write("(")
        if args:
            self.cast_null(args[0])
        for arg in args[1:]:
            self.write(', ')
            self.cast_null(arg)
        self.write(")")

        if optargs:
            self.to_args_optargs(func, optargs)

    def to_args_optargs(self, func='', optargs=[]):
        optarg_aliases = {
            'JsOpts': 'JSOpts',
            'HttpOpts': 'HTTPOpts',
            'Iso8601Opts': 'ISO8601Opts',
            'IndexCreateFuncOpts': 'IndexCreateOpts'
        }
        optarg_field_aliases = {
            'nonvoting_replica_tags': 'NonVotingReplicaTags',
        }

        if not func:
            raise Unhandled("Missing function name")

        optarg_type = camel(func) + 'Opts'
        optarg_type = optarg_aliases.get(optarg_type, optarg_type)
        optarg_type = 'r.' + optarg_type

        self.write('.OptArgs(')
        self.write(optarg_type)
        self.write('{')
        for optarg in optargs:
            # Hack to skip tests that check for unknown opt args,
            # this is not possible in Go due to static types
            if optarg.arg == 'return_vals':
                self.skip("test not required since optargs are statically typed")
                return
            if optarg.arg == 'foo':
                self.skip("test not required since optargs are statically typed")
                return
            if type(optarg.value) == ast.Name and optarg.value.id == 'null':
                self.skip("test not required since go does not support null optargs")
                return


            field_name = optarg_field_aliases.get(optarg.arg, camel(optarg.arg))

            self.write(field_name)
            self.write(": ")
            self.cast_null(optarg.value)
            self.write(', ')
        self.write('})')

    def generic_visit(self, node):
        logger.error("While translating: %s", ast.dump(node))
        logger.error("Got as far as: %s", ''.join(self.out))
        raise Unhandled("Don't know what this thing is: " + str(type(node)))

    def visit_Assign(self, node):
        if len(node.targets) != 1:
            Unhandled("We only support assigning to one variable")
        var = node.targets[0].id
        self.write("var " + var + " ")
        if is_reql(self._type):
            self.write('r.Term')
        else:
            self.write(self.type)
        self.write(" = ")
        if is_reql(self._type):
            ReQLVisitor(self.reql_vars,
                        out=self.out,
                        type_=self.type,
                        is_def=True,
                        ).visit(node.value)
        elif var == 'upper_limit': # Manually set value since value in test causes an error
            self.write('2<<52 - 1')
        elif var == 'lower_limit': # Manually set value since value in test causes an error
            self.write('1 - 2<<52')
        else:
            self.visit(node.value)

    def visit_Str(self, node):
        if node.s == 'ReqlServerCompileError':
            node.s = 'ReqlCompileError'
        # Hack to skip irrelevant tests
        if match(".*Expected .* argument", node.s):
            self.skip("argument checks not supported")
        if match(".*argument .* must", node.s):
            self.skip("argument checks not supported")
            return
        if node.s == 'Object keys must be strings.*':
            self.skip('the Go driver automatically converts object keys to strings')
            return
        if node.s.startswith('\'module\' object has no attribute '):
            self.skip('test not required since terms are statically typed')
            return
        self.to_str(node.s)

    def visit_Bytes(self, node, skip_prefix=False, skip_suffix=False):
        if not skip_prefix:
            self.write("[]byte{")
        for i, byte in enumerate(node.s):
            self.write(str(byte))
            if i < len(node.s)-1:
                self.write(",")
        if not skip_suffix:
            self.write("}")
        else:
            self.write(", ")

    def visit_Name(self, node):
        name = node.id
        if name == 'frozenset':
            self.skip("can't convert frozensets to GroupedData yet")
        if name in GO_KEYWORDS:
            name += '_'
        self.write({
            'True': 'true',
            'False': 'false',
            'None': 'nil',
            'nil': 'nil',
            'null': 'nil',
            'float': 'float64',
            'datetime': 'Ast', # Use helper method instead
            'len': 'maybeLen',
            'AnythingIsFine': 'compare.AnythingIsFine',
            'bag': 'compare.UnorderedMatch',
            'partial': 'compare.PartialMatch',
            'uuid': 'compare.IsUUID',
            'regex': 'compare.MatchesRegexp',
            }.get(name, name))

    def visit_arg(self, node):
        self.write(node.arg)
        self.write(" ")
        if is_reql(self._type):
            self.write("r.Term")
        else:
            self.write("interface{}")

    def visit_NameConstant(self, node):
        if node.value is None:
            self.write("nil")
        elif node.value is True:
            self.write("true")
        elif node.value is False:
            self.write("false")
        else:
            raise Unhandled(
                "Don't know NameConstant with value %s" % node.value)

    def visit_Attribute(self, node, emit_parens=True):
        skip_parent = False
        if attr_matches("r.ast", node):
            # The java driver doesn't have that namespace, so we skip
            # the `r.` prefix and create an ast class member in the
            # test file. So stuff like `r.ast.rqlTzinfo(...)` converts
            # to `ast.rqlTzinfo(...)`
            skip_parent = True


        if not skip_parent:
            self.visit(node.value)
            self.write(".")

        attr = ReQLVisitor.convertTermName(node.attr)
        self.write(attr)

    def visit_Num(self, node):
        self.write(repr(node.n))
        if not isinstance(node.n, float):
            if node.n > 9223372036854775807 or node.n < -9223372036854775808:
                self.write(".0")

    def visit_Index(self, node):
        self.visit(node.value)

    def visit_Call(self, node):
        func = ''
        if type(node.func) == ast.Attribute and node.func.attr == 'encode':
            self.skip("Go tests do not currently support character encoding")
            return
        if type(node.func) == ast.Attribute and node.func.attr == 'error':
            # This weird special case is because sometimes the tests
            # use r.error and sometimes they use r.error(). The java
            # driver only supports r.error(). Since we're coming in
            # from a call here, we have to prevent visit_Attribute
            # from emitting the parents on an r.error for us.
            self.visit_Attribute(node.func, emit_parens=False)
        elif type(node.func) == ast.Name and node.func.id == 'err':
            # Throw away third argument as it is not used by the Go tests
            # and Go does not support function overloading
            node.args = node.args[:2]
            self.visit(node.func)
        elif type(node.func) == ast.Name and node.func.id == 'err_regex':
            # Throw away third argument as it is not used by the Go tests
            # and Go does not support function overloading
            node.args = node.args[:2]
            self.visit(node.func)
        elif type(node.func) == ast.Name and node.func.id == 'fetch':
            if len(node.args) == 1:
                node.args.append(ast.Num(0))
            elif len(node.args) > 2:
                node.args = node.args[:2]
            self.visit(node.func)
        else:
            self.visit(node.func)

        if type(node.func) == ast.Attribute:
            func = node.func.attr
        elif type(node.func) == ast.Name:
            func = node.func.id

        self.to_args(node.args, func, node.keywords)

    def visit_Dict(self, node):
        self.write("map[interface{}]interface{}{")
        for k, v in zip(node.keys, node.values):
            self.visit(k)
            self.write(": ")
            self.visit(v)
            self.write(", ")
        self.write("}")

    def visit_List(self, node):
        self.write("[]interface{}{")
        self.join(", ", node.elts)
        self.write("}")

    def visit_Tuple(self, node):
        self.visit_List(node)

    def visit_Lambda(self, node):
        self.write("func")
        self.to_args(node.args.args)
        self.write(" interface{} { return ")
        self.visit(node.body)
        self.write("}")

    def visit_Subscript(self, node):
        if node.slice is None or type(node.slice.value) != ast.Num:
            logger.error("While doing: %s", ast.dump(node))
            raise Unhandled("Only integers subscript can be converted."
                            " Got %s" % node.slice.value.s)
        self.visit(node.value)
        self.write("[")
        self.write(str(node.slice.value.n))
        self.write("]")

    def visit_ListComp(self, node):
        gen = node.generators[0]

        start = 0
        end = 0

        if type(gen.iter) == ast.Call and gen.iter.func.id.endswith('range'):
            # This is really a special-case hacking of [... for i in
            # range(i)] comprehensions that are used in the polyglot
            # tests sometimes. It won't handle translating arbitrary
            # comprehensions to Java streams.

            if len(gen.iter.args) == 1:
                end = gen.iter.args[0].n
            elif len(gen.iter.args) == 2:
                start = gen.iter.args[0].n
                end = gen.iter.args[1].n
        else:
            # Somebody came up with a creative new use for
            # comprehensions in the test suite...
            raise Unhandled("ListComp hack couldn't handle: ", ast.dump(node))

        self.write("(func() []interface{} {\n")
        self.write("    res := []interface{}{}\n")
        self.write("    for iterator_ := %s; iterator_ < %s; iterator_++ {\n" % (start, end))
        self.write("        "); self.visit(gen.target); self.write(" := iterator_\n")
        self.write("        res = append(res, ")
        self.visit(node.elt)
        self.write(")\n")
        self.write("    }\n")
        self.write("    return res\n")
        self.write("}())")

    def visit_UnaryOp(self, node):
        opMap = {
            ast.USub: "-",
            ast.Not: "!",
            ast.UAdd: "+",
            ast.Invert: "~",
        }
        self.write(opMap[type(node.op)])
        self.visit(node.operand)

    def visit_BinOp(self, node):
        opMap = {
            ast.Add: " + ",
            ast.Sub: " - ",
            ast.Mult: " * ",
            ast.Div: " / ",
            ast.Mod: " % ",
        }
        if self.is_string_mul(node):
            return
        if self.is_array_concat(node):
            return
        if self.is_byte_array_add(node):
            return

        t = type(node.op)
        if t in opMap.keys():
            self.visit(node.left)
            self.write(opMap[t])
            self.visit(node.right)
        elif t == ast.Pow:
            if type(node.left) == ast.Num and node.left.n == 2:
                self.visit(node.left)
                self.write(" << ")
                self.visit(node.right)
            else:
                raise Unhandled("Can't do exponent with non 2 base")

    def is_byte_array_add(self, node):
        '''Some places we do stuff like b'foo' + b'bar' and byte
        arrays don't like that much'''
        if (type(node.left) == ast.Bytes and
           type(node.right) == ast.Bytes and
           type(node.op) == ast.Add):
            self.visit_Bytes(node.left, skip_suffix=True)
            self.visit_Bytes(node.right, skip_prefix=True)
            return True
        else:
            return False

    def is_string_mul(self, node):
        if ((type(node.left) == ast.Str and type(node.right) == ast.Num) and type(node.op) == ast.Mult):
            self.write("\"")
            self.write(node.left.s * node.right.n)
            self.write("\"")
            return True
        elif ((type(node.left) == ast.Num and type(node.right) == ast.Str) and type(node.op) == ast.Mult):
            self.write("\"")
            self.write(node.left.n * node.right.s)
            self.write("\"")
            return True
        else:
            return False

    def is_array_concat(self, node):
        if ((type(node.left) == ast.List or type(node.right) == ast.List) and type(node.op) == ast.Add):
            self.skip("Array concatenation using + operator not currently supported")
            return True
        else:
            return False


class ReQLVisitor(GoVisitor):
    '''Mostly the same as the GoVisitor, but converts some
    reql-specific stuff. This should only be invoked on an expression
    if it's already known to return true from is_reql'''

    TOPLEVEL_CONSTANTS = {
        'error'
    }

    def visit_BinOp(self, node):
        if self.is_string_mul(node):
            return
        if self.is_array_concat(node):
            return
        if self.is_byte_array_add(node):
            return
        opMap = {
            ast.Add: "Add",
            ast.Sub: "Sub",
            ast.Mult: "Mul",
            ast.Div: "Div",
            ast.Mod: "Mod",
            ast.BitAnd: "And",
            ast.BitOr: "Or",
        }
        func = opMap[type(node.op)]
        if self.is_not_reql(node.left):
            self.prefix(func, node.left, node.right)
        else:
            self.infix(func, node.left, node.right)

    def visit_Compare(self, node):
        opMap = {
            ast.Lt: "Lt",
            ast.Gt: "Gt",
            ast.GtE: "Ge",
            ast.LtE: "Le",
            ast.Eq: "Eq",
            ast.NotEq: "Ne",
        }
        if len(node.ops) != 1:
            # Python syntax allows chained comparisons (a < b < c) but
            # we don't deal with that here
            raise Unhandled("Compare hack bailed on: ", ast.dump(node))
        left = node.left
        right = node.comparators[0]
        func_name = opMap[type(node.ops[0])]
        if self.is_not_reql(node.left):
            self.prefix(func_name, left, right)
        else:
            self.infix(func_name, left, right)

    def prefix(self, func_name, left, right):
        self.write("r.")
        self.write(func_name)
        self.write("(")
        self.visit(left)
        self.write(", ")
        self.visit(right)
        self.write(")")

    def infix(self, func_name, left, right):
        self.visit(left)
        self.write(".")
        self.write(func_name)
        self.write("(")
        self.visit(right)
        self.write(")")

    def is_not_reql(self, node):
        if type(node) in (ast.Name, ast.NameConstant,
                          ast.Num, ast.Str, ast.Dict, ast.List):
            return True
        else:
            return False

    def visit_Subscript(self, node):
        self.visit(node.value)
        if type(node.slice) == ast.Index:
            # Syntax like a[2] or a["b"]
            if self.smart_bracket and type(node.slice.value) == ast.Str:
                self.write(".Field(")
            elif self.smart_bracket and type(node.slice.value) == ast.Num:
                self.write(".Nth(")
            else:
                self.write(".AtIndex(")
            self.visit(node.slice.value)
            self.write(")")
        elif type(node.slice) == ast.Slice:
            # Syntax like a[1:2] or a[:2]
            self.write(".Slice(")
            lower, upper, rclosed = self.get_slice_bounds(node.slice)
            self.write(str(lower))
            self.write(", ")
            self.write(str(upper))
            if rclosed:
                self.write(', r.SliceOpts{RightBound: "closed"}')
            self.write(")")
        else:
            raise Unhandled("No translation for ExtSlice")

    def get_slice_bounds(self, slc):
        '''Used to extract bounds when using bracket slice
        syntax. This is more complicated since Python3 parses -1 as
        UnaryOp(op=USub, operand=Num(1)) instead of Num(-1) like
        Python2 does'''
        if not slc:
            return 0, -1, True

        def get_bound(bound, default):
            if bound is None:
                return default
            elif type(bound) == ast.UnaryOp and type(bound.op) == ast.USub:
                return -bound.operand.n
            elif type(bound) == ast.Num:
                return bound.n
            else:
                raise Unhandled(
                    "Not handling bound: %s" % ast.dump(bound))

        right_closed = slc.upper is None

        return get_bound(slc.lower, 0), get_bound(slc.upper, -1), right_closed

    def convertTermName(term):
        python_clashes = {
            # These are underscored in the python driver to avoid
            # keywords, but they aren't java keywords so we convert
            # them back.
            'or_': 'Or',
            'and_': 'And',
            'not_': 'Not',
        }
        method_aliases = {
            'get_field': 'Field',
            'db': 'DB',
            'db_create': 'DBCreate',
            'db_drop': 'DBDrop',
            'db_list': 'DBList',
            'uuid': 'UUID',
            'geojson': 'GeoJSON',
            'js': 'JS',
            'json': 'JSON',
            'to_json': 'ToJSON',
            'to_json_string': 'ToJSON',
            'minval': 'MinVal',
            'maxval': 'MaxVal',
            'http': 'HTTP',
            'iso8601': 'ISO8601',
            'to_iso8601': 'ToISO8601',
        }

        return python_clashes.get(term, method_aliases.get(term, camel(term)))

    def visit_Attribute(self, node, emit_parens=True):
        is_toplevel_constant = False
        # if attr_matches("r.row", node):
        # elif is_name("r", node.value) and node.attr in self.TOPLEVEL_CONSTANTS:
        if is_name("r", node.value) and node.attr in self.TOPLEVEL_CONSTANTS:
            # Python has r.minval, r.saturday etc. We need to emit
            # r.minval() and r.saturday()
            is_toplevel_constant = True

        initial = ReQLVisitor.convertTermName(node.attr)

        self.visit(node.value)
        self.write(".")
        self.write(initial)
        if initial in GO_KEYWORDS:
            self.write('_')
        if emit_parens and is_toplevel_constant:
            self.write('()')

    def visit_UnaryOp(self, node):
        if type(node.op) == ast.Invert:
            self.visit(node.operand)
            self.write(".Not()")
        else:
            super(ReQLVisitor, self).visit_UnaryOp(node)

    def visit_Call(self, node):
        if (attr_equals(node.func, "attr", "index_create") and len(node.args) == 2):
            node.func.attr = 'index_create_func'

        # We call the superclass first, so if it's going to fail
        # because of r.row or other things it fails first, rather than
        # hitting the checks in this method. Since everything is
        # written to a stringIO object not directly to a file, if we
        # bail out afterwards it's still ok
        super_result = super(ReQLVisitor, self).visit_Call(node)

        # r.expr(v, 1) should be skipped
        if (attr_equals(node.func, "attr", "expr") and len(node.args) > 1):
            self.skip("the go driver only accepts one parameter to expr")
        # r.table_create("a", "b") should be skipped
        if (attr_equals(node.func, "attr", "table_create") and len(node.args) > 1):
            self.skip("the go driver only accepts one parameter to table_create")
        return super_result

class Renderer(object):
    '''Manages rendering templates'''

    def __init__(self, invoking_filenames, source_files=None):
        self.template_file = './template.go.tpl'
        self.invoking_filenames = invoking_filenames
        self.source_files = source_files or []
        self.tpl = Template(filename=self.template_file)
        self.template_context = {
            'EmptyTemplate': process_polyglot.EmptyTemplate,
        }

    def render(self,
               template_name,
               output_dir,
               output_name=None,
               **kwargs):
        if output_name is None:
            output_name = template_name

        output_path = output_dir + '/' + output_name

        results = self.template_context.copy()
        results.update(kwargs)
        try:
            rendered = self.tpl.render(**results)
        except process_polyglot.EmptyTemplate:
            logger.debug(" Empty template: %s", output_path)
            return
        with codecs.open(output_path, "w", "utf-8") as outfile:
            logger.info("Rendering %s", output_path)
            outfile.write(self.autogenerated_header(
                self.template_file,
                output_path,
                self.invoking_filenames,
            ))
            outfile.write(rendered)

    def autogenerated_header(self, template_path, output_path, filename):
        rel_tpl = os.path.relpath(template_path, start=output_path)
        filenames = ' and '.join(os.path.basename(f)
                                 for f in self.invoking_filenames)
        return ('// Code generated by {}.\n'
                '// Do not edit this file directly.\n'
                '// The template for this file is located at:\n'
                '// {}\n').format(filenames, rel_tpl)

def camel(varname):
    'CamelCase'
    if re.match(r'[A-Z][A-Z0-9_]*$|[a-z][a-z0-9_]*$', varname):
        # if snake-case (upper or lower) camelize it
        suffix = "_" if varname.endswith('_') else ""
        return ''.join(x.title() for x in varname.split('_')) + suffix
    else:
        # if already mixed case, just capitalize the first letter
        return varname[0].upper() + varname[1:]


def dromedary(varname):
    'dromedaryCase'
    if re.match(r'[A-Z][A-Z0-9_]*$|[a-z][a-z0-9_]*$', varname):
        chunks = varname.split('_')
        suffix = "_" if varname.endswith('_') else ""
        return (chunks[0].lower() +
                ''.join(x.title() for x in chunks[1:]) +
                suffix)
    else:
        return varname[0].lower() + varname[1:]

def attr_equals(node, attr, value):
    '''Helper for digging into ast nodes'''
    return hasattr(node, attr) and getattr(node, attr) == value

if __name__ == '__main__':
    main()
