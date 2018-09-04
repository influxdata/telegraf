package pgmock

import (
	"io"
	"net"
	"reflect"

	"github.com/pkg/errors"

	"github.com/jackc/pgx/pgproto3"
	"github.com/jackc/pgx/pgtype"
)

type Server struct {
	ln         net.Listener
	controller Controller
}

func NewServer(controller Controller) (*Server, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}

	server := &Server{
		ln:         ln,
		controller: controller,
	}

	return server, nil
}

func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *Server) ServeOne() error {
	conn, err := s.ln.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	s.Close()

	backend, err := pgproto3.NewBackend(conn, conn)
	if err != nil {
		conn.Close()
		return err
	}

	return s.controller.Serve(backend)
}

func (s *Server) Close() error {
	err := s.ln.Close()
	if err != nil {
		return err
	}

	return nil
}

type Controller interface {
	Serve(backend *pgproto3.Backend) error
}

type Step interface {
	Step(*pgproto3.Backend) error
}

type Script struct {
	Steps []Step
}

func (s *Script) Run(backend *pgproto3.Backend) error {
	for _, step := range s.Steps {
		err := step.Step(backend)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Script) Serve(backend *pgproto3.Backend) error {
	for _, step := range s.Steps {
		err := step.Step(backend)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Script) Step(backend *pgproto3.Backend) error {
	return s.Serve(backend)
}

type expectMessageStep struct {
	want pgproto3.FrontendMessage
	any  bool
}

func (e *expectMessageStep) Step(backend *pgproto3.Backend) error {
	msg, err := backend.Receive()
	if err != nil {
		return err
	}

	if e.any && reflect.TypeOf(msg) == reflect.TypeOf(e.want) {
		return nil
	}

	if !reflect.DeepEqual(msg, e.want) {
		return errors.Errorf("msg => %#v, e.want => %#v", msg, e.want)
	}

	return nil
}

type expectStartupMessageStep struct {
	want *pgproto3.StartupMessage
	any  bool
}

func (e *expectStartupMessageStep) Step(backend *pgproto3.Backend) error {
	msg, err := backend.ReceiveStartupMessage()
	if err != nil {
		return err
	}

	if e.any {
		return nil
	}

	if !reflect.DeepEqual(msg, e.want) {
		return errors.Errorf("msg => %#v, e.want => %#v", msg, e.want)
	}

	return nil
}

func ExpectMessage(want pgproto3.FrontendMessage) Step {
	return expectMessage(want, false)
}

func ExpectAnyMessage(want pgproto3.FrontendMessage) Step {
	return expectMessage(want, true)
}

func expectMessage(want pgproto3.FrontendMessage, any bool) Step {
	if want, ok := want.(*pgproto3.StartupMessage); ok {
		return &expectStartupMessageStep{want: want, any: any}
	}

	return &expectMessageStep{want: want, any: any}
}

type sendMessageStep struct {
	msg pgproto3.BackendMessage
}

func (e *sendMessageStep) Step(backend *pgproto3.Backend) error {
	return backend.Send(e.msg)
}

func SendMessage(msg pgproto3.BackendMessage) Step {
	return &sendMessageStep{msg: msg}
}

type waitForCloseMessageStep struct{}

func (e *waitForCloseMessageStep) Step(backend *pgproto3.Backend) error {
	for {
		msg, err := backend.Receive()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if _, ok := msg.(*pgproto3.Terminate); ok {
			return nil
		}
	}
}

func WaitForClose() Step {
	return &waitForCloseMessageStep{}
}

func AcceptUnauthenticatedConnRequestSteps() []Step {
	return []Step{
		ExpectAnyMessage(&pgproto3.StartupMessage{ProtocolVersion: pgproto3.ProtocolVersionNumber, Parameters: map[string]string{}}),
		SendMessage(&pgproto3.Authentication{Type: pgproto3.AuthTypeOk}),
		SendMessage(&pgproto3.BackendKeyData{ProcessID: 0, SecretKey: 0}),
		SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'I'}),
	}
}

func PgxInitSteps() []Step {
	steps := []Step{
		ExpectMessage(&pgproto3.Parse{
			Query: `select t.oid,
	case when nsp.nspname in ('pg_catalog', 'public') then t.typname
		else nsp.nspname||'.'||t.typname
	end
from pg_type t
left join pg_type base_type on t.typelem=base_type.oid
left join pg_namespace nsp on t.typnamespace=nsp.oid
where (
	  t.typtype in('b', 'p', 'r', 'e')
	  and (base_type.oid is null or base_type.typtype in('b', 'p', 'r'))
	)`,
		}),
		ExpectMessage(&pgproto3.Describe{
			ObjectType: 'S',
		}),
		ExpectMessage(&pgproto3.Sync{}),
		SendMessage(&pgproto3.ParseComplete{}),
		SendMessage(&pgproto3.ParameterDescription{}),
		SendMessage(&pgproto3.RowDescription{
			Fields: []pgproto3.FieldDescription{
				{Name: "oid",
					TableOID:             1247,
					TableAttributeNumber: 65534,
					DataTypeOID:          26,
					DataTypeSize:         4,
					TypeModifier:         4294967295,
					Format:               0,
				},
				{Name: "typname",
					TableOID:             1247,
					TableAttributeNumber: 1,
					DataTypeOID:          19,
					DataTypeSize:         64,
					TypeModifier:         4294967295,
					Format:               0,
				},
			},
		}),
		SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'I'}),
		ExpectMessage(&pgproto3.Bind{
			ResultFormatCodes: []int16{1, 1},
		}),
		ExpectMessage(&pgproto3.Execute{}),
		ExpectMessage(&pgproto3.Sync{}),
		SendMessage(&pgproto3.BindComplete{}),
	}

	rowVals := []struct {
		oid  pgtype.OID
		name string
	}{
		{16, "bool"},
		{17, "bytea"},
		{18, "char"},
		{19, "name"},
		{20, "int8"},
		{21, "int2"},
		{22, "int2vector"},
		{23, "int4"},
		{24, "regproc"},
		{25, "text"},
		{26, "oid"},
		{27, "tid"},
		{28, "xid"},
		{29, "cid"},
		{30, "oidvector"},
		{114, "json"},
		{142, "xml"},
		{143, "_xml"},
		{199, "_json"},
		{194, "pg_node_tree"},
		{32, "pg_ddl_command"},
		{210, "smgr"},
		{600, "point"},
		{601, "lseg"},
		{602, "path"},
		{603, "box"},
		{604, "polygon"},
		{628, "line"},
		{629, "_line"},
		{700, "float4"},
		{701, "float8"},
		{702, "abstime"},
		{703, "reltime"},
		{704, "tinterval"},
		{705, "unknown"},
		{718, "circle"},
		{719, "_circle"},
		{790, "money"},
		{791, "_money"},
		{829, "macaddr"},
		{869, "inet"},
		{650, "cidr"},
		{1000, "_bool"},
		{1001, "_bytea"},
		{1002, "_char"},
		{1003, "_name"},
		{1005, "_int2"},
		{1006, "_int2vector"},
		{1007, "_int4"},
		{1008, "_regproc"},
		{1009, "_text"},
		{1028, "_oid"},
		{1010, "_tid"},
		{1011, "_xid"},
		{1012, "_cid"},
		{1013, "_oidvector"},
		{1014, "_bpchar"},
		{1015, "_varchar"},
		{1016, "_int8"},
		{1017, "_point"},
		{1018, "_lseg"},
		{1019, "_path"},
		{1020, "_box"},
		{1021, "_float4"},
		{1022, "_float8"},
		{1023, "_abstime"},
		{1024, "_reltime"},
		{1025, "_tinterval"},
		{1027, "_polygon"},
		{1033, "aclitem"},
		{1034, "_aclitem"},
		{1040, "_macaddr"},
		{1041, "_inet"},
		{651, "_cidr"},
		{1263, "_cstring"},
		{1042, "bpchar"},
		{1043, "varchar"},
		{1082, "date"},
		{1083, "time"},
		{1114, "timestamp"},
		{1115, "_timestamp"},
		{1182, "_date"},
		{1183, "_time"},
		{1184, "timestamptz"},
		{1185, "_timestamptz"},
		{1186, "interval"},
		{1187, "_interval"},
		{1231, "_numeric"},
		{1266, "timetz"},
		{1270, "_timetz"},
		{1560, "bit"},
		{1561, "_bit"},
		{1562, "varbit"},
		{1563, "_varbit"},
		{1700, "numeric"},
		{1790, "refcursor"},
		{2201, "_refcursor"},
		{2202, "regprocedure"},
		{2203, "regoper"},
		{2204, "regoperator"},
		{2205, "regclass"},
		{2206, "regtype"},
		{4096, "regrole"},
		{4089, "regnamespace"},
		{2207, "_regprocedure"},
		{2208, "_regoper"},
		{2209, "_regoperator"},
		{2210, "_regclass"},
		{2211, "_regtype"},
		{4097, "_regrole"},
		{4090, "_regnamespace"},
		{2950, "uuid"},
		{2951, "_uuid"},
		{3220, "pg_lsn"},
		{3221, "_pg_lsn"},
		{3614, "tsvector"},
		{3642, "gtsvector"},
		{3615, "tsquery"},
		{3734, "regconfig"},
		{3769, "regdictionary"},
		{3643, "_tsvector"},
		{3644, "_gtsvector"},
		{3645, "_tsquery"},
		{3735, "_regconfig"},
		{3770, "_regdictionary"},
		{3802, "jsonb"},
		{3807, "_jsonb"},
		{2970, "txid_snapshot"},
		{2949, "_txid_snapshot"},
		{3904, "int4range"},
		{3905, "_int4range"},
		{3906, "numrange"},
		{3907, "_numrange"},
		{3908, "tsrange"},
		{3909, "_tsrange"},
		{3910, "tstzrange"},
		{3911, "_tstzrange"},
		{3912, "daterange"},
		{3913, "_daterange"},
		{3926, "int8range"},
		{3927, "_int8range"},
		{2249, "record"},
		{2287, "_record"},
		{2275, "cstring"},
		{2276, "any"},
		{2277, "anyarray"},
		{2278, "void"},
		{2279, "trigger"},
		{3838, "event_trigger"},
		{2280, "language_handler"},
		{2281, "internal"},
		{2282, "opaque"},
		{2283, "anyelement"},
		{2776, "anynonarray"},
		{3500, "anyenum"},
		{3115, "fdw_handler"},
		{325, "index_am_handler"},
		{3310, "tsm_handler"},
		{3831, "anyrange"},
		{51367, "gbtreekey4"},
		{51370, "_gbtreekey4"},
		{51371, "gbtreekey8"},
		{51374, "_gbtreekey8"},
		{51375, "gbtreekey16"},
		{51378, "_gbtreekey16"},
		{51379, "gbtreekey32"},
		{51382, "_gbtreekey32"},
		{51383, "gbtreekey_var"},
		{51386, "_gbtreekey_var"},
		{51921, "hstore"},
		{51926, "_hstore"},
		{52005, "ghstore"},
		{52008, "_ghstore"},
	}

	for _, rv := range rowVals {
		step := SendMessage(mustBuildDataRow([]interface{}{rv.oid, rv.name}, []int16{pgproto3.BinaryFormat}))
		steps = append(steps, step)
	}

	steps = append(steps, SendMessage(&pgproto3.CommandComplete{CommandTag: "SELECT 163"}))
	steps = append(steps, SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'I'}))

	steps = append(steps, []Step{
		ExpectMessage(&pgproto3.Parse{
			Query: "select t.oid, t.typname\nfrom pg_type t\n  join pg_type base_type on t.typelem=base_type.oid\nwhere t.typtype = 'b'\n  and base_type.typtype = 'e'",
		}),
		ExpectMessage(&pgproto3.Describe{
			ObjectType: 'S',
		}),
		ExpectMessage(&pgproto3.Sync{}),
		SendMessage(&pgproto3.ParseComplete{}),
		SendMessage(&pgproto3.ParameterDescription{}),
		SendMessage(&pgproto3.RowDescription{
			Fields: []pgproto3.FieldDescription{
				{Name: "oid",
					TableOID:             1247,
					TableAttributeNumber: 65534,
					DataTypeOID:          26,
					DataTypeSize:         4,
					TypeModifier:         4294967295,
					Format:               0,
				},
				{Name: "typname",
					TableOID:             1247,
					TableAttributeNumber: 1,
					DataTypeOID:          19,
					DataTypeSize:         64,
					TypeModifier:         4294967295,
					Format:               0,
				},
			},
		}),
		SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'I'}),
		ExpectMessage(&pgproto3.Bind{
			ResultFormatCodes: []int16{1, 1},
		}),
		ExpectMessage(&pgproto3.Execute{}),
		ExpectMessage(&pgproto3.Sync{}),
		SendMessage(&pgproto3.BindComplete{}),
		SendMessage(&pgproto3.CommandComplete{CommandTag: "SELECT 0"}),
		SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'I'}),
	}...)

	return steps
}

type dataRowValue struct {
	Value      interface{}
	FormatCode int16
}

func mustBuildDataRow(values []interface{}, formatCodes []int16) *pgproto3.DataRow {
	dr, err := buildDataRow(values, formatCodes)
	if err != nil {
		panic(err)
	}

	return dr
}

func buildDataRow(values []interface{}, formatCodes []int16) (*pgproto3.DataRow, error) {
	dr := &pgproto3.DataRow{
		Values: make([][]byte, len(values)),
	}

	if len(formatCodes) == 1 {
		for i := 1; i < len(values); i++ {
			formatCodes = append(formatCodes, formatCodes[0])
		}
	}

	for i := range values {
		switch v := values[i].(type) {
		case string:
			values[i] = &pgtype.Text{String: v, Status: pgtype.Present}
		case int16:
			values[i] = &pgtype.Int2{Int: v, Status: pgtype.Present}
		case int32:
			values[i] = &pgtype.Int4{Int: v, Status: pgtype.Present}
		case int64:
			values[i] = &pgtype.Int8{Int: v, Status: pgtype.Present}
		}
	}

	for i := range values {
		switch formatCodes[i] {
		case pgproto3.TextFormat:
			if e, ok := values[i].(pgtype.TextEncoder); ok {
				buf, err := e.EncodeText(nil, nil)
				if err != nil {
					return nil, errors.Errorf("failed to encode values[%d]", i)
				}
				dr.Values[i] = buf
			} else {
				return nil, errors.Errorf("values[%d] does not implement TextExcoder", i)
			}

		case pgproto3.BinaryFormat:
			if e, ok := values[i].(pgtype.BinaryEncoder); ok {
				buf, err := e.EncodeBinary(nil, nil)
				if err != nil {
					return nil, errors.Errorf("failed to encode values[%d]", i)
				}
				dr.Values[i] = buf
			} else {
				return nil, errors.Errorf("values[%d] does not implement BinaryEncoder", i)
			}
		default:
			return nil, errors.New("unknown FormatCode")
		}
	}

	return dr, nil
}
