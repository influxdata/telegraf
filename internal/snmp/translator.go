package snmp

type TranslatorPlugin interface {
	SetTranslator(name string) // Agent calls this on inputs before Init
}

type Translator interface {
	SnmpTranslate(oid string) (
		mibName string, oidNum string, oidText string,
		conversion string,
		err error,
	)

	SnmpTable(oid string) (
		mibName string, oidNum string, oidText string,
		fields []Field,
		err error,
	)

	SnmpFormatEnum(oid string, value interface{}, full bool) (
		formatted string,
		err error,
	)
}
