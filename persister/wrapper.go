package persister

type PersisterPluginWrapper struct {
	id        string     // plugin instance ID
	persister *Persister // underlying persister instance
}

func (w *PersisterPluginWrapper) UpdateState(state interface{}) error {
	return w.persister.SetState(w.id, state)
}
