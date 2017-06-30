package zipkin

/*func TestMainHandler(t *testing.T) {
	dat, err := ioutil.ReadFile("test/threespans.dat")
	if err != nil {
		t.Fatal("Could not find threespans.dat")
	}
	e := make(chan error, 1)
	d := make(chan SpanData, 1)
	//f := make(chan struct{}, 1)
	s := NewHTTPServer(9411, e, d)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"POST",
		"http://server.local/api/v1/spans",
		ioutil.NopCloser(
			bytes.NewReader(dat)))
	handler := s.MainHandler()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Errorf("MainHandler did not return StatusNoContent %d", w.Code)
	}

	spans := <-d
	if len(spans) != 3 {
		t.Fatalf("Expected 3 spans received len(spans) %d", len(spans))
	}
	if spans[0].ID != 8090652509916334619 {
		t.Errorf("Expected 8090652509916334619 but received ID %d ", spans[0].ID)
	}
	if spans[0].TraceID != 2505404965370368069 {
		t.Errorf("Expected 2505404965370368069 but received TraceID %d ", spans[0].TraceID)
	}
	if spans[0].Name != "Child" {
		t.Errorf("Expected Child but recieved name %s", spans[0].Name)
	}
	if *(spans[0].ParentID) != 22964302721410078 {
		t.Errorf("Expected 22964302721410078 but recieved parent id %d", spans[0].ParentID)
	}
	serviceName := spans[0].GetBinaryAnnotations()[0].GetHost().GetServiceName()
	if serviceName != "trivial" {
		t.Errorf("Expected trivial but recieved service name %s", serviceName)
	}

	if spans[0].GetTimestamp() != 1498688360851331 {
		t.Errorf("Expected timestamp %d but recieved timestamp %d", 1498688360851331, spans[0].GetTimestamp())
	}

	if spans[0].GetDuration() != 53106 {
		t.Errorf("Expected duration %d but recieved duration %d", 53106, spans[0].GetDuration())
	}
}*/
