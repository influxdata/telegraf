


[IBM ZID file](https://www.ibm.com/support/knowledgecenter/en/SSVRGU_8.5.3/com.ibm.designer.domino.main.doc/H_FIELD_OPTIONS_AND_TEXT_INFORMATION_DEFINITION_SYNTAX_2922_OVER.html)

[IBM ZID Example](https://www.ibm.com/support/knowledgecenter/en/SSVRGU_10.0.0/basic/H_BINARY_INPUT_FILES_USING_FIXED_LENGTH_RECORDS_6416_OVER.html)



```toml
[fixed_record]
  metric_name = "drone_status"
  endiannes = "be"

  fields = [
    {name="version",type="uint16",offset=0},
    {name="time",type="int32",offset=2},
  ]

```

