package snmp

type TranslatorPlugin interface {
	SetTranslator(name string) // Agent calls this on inputs before Init
}
