# Windows Eventlog Input Plugin

## Collect Windows Event Log messages

Supports Windows Vista and higher.

Telegraf should have Administrator permissions to subscribe for Windows Events.

### Configuration

```toml
  ## LCID (Locale ID) for event rendering
  ## 1033 to force English language
  ## 0 to use default Windows locale
  # locale = 0

  ## Name of eventlog, used only if xpath_query is empty
  ## Example: "Application"
  eventlog_name = ""

  ## xpath_query can be in defined short form like "Event/System[EventID=999]"
  ## or you can form a xml query. Refer to the Consuming Events article:
  ## https://docs.microsoft.com/en-us/windows/win32/wes/consuming-events
  xpath_query = '''
  <QueryList>
    <Query Id="0" Path="Security">
      <Select Path="Security">*</Select>
      <Suppress Path="Security">*[System[( (EventID &gt;= 5152 and EventID &lt;= 5158) or EventID=5379 or EventID=4672)]]</Suppress>
    </Query>
    <Query Id="1" Path="Application">
      <Select Path="Application">*[System[(Level &lt; 4)]]</Select>
      <Select Path="OpenSSH/Admin">*[System[(Level &lt; 4)]]</Select>
      <Select Path="Windows PowerShell">*[System[(Level &lt; 4)]]</Select>
      <Select Path="Key Management Service">*[System[(Level &lt; 4)]]</Select>
      <Select Path="HardwareEvents">*[System[(Level &lt; 4)]]</Select>
    </Query>
    <Query Id="2" Path="Windows PowerShell">
      <Select Path="Windows PowerShell">*[System[(Level &lt; 4)]]</Select>
    </Query>
    <Query Id="3" Path="System">
      <Select Path="System">*</Select>
    </Query>
    <Query Id="4" Path="Setup">
      <Select Path="Setup">*</Select>
    </Query>
  </QueryList>
  '''
```

### Filtering

There are three types of filtering: *Event Log* name, *XPath Query* and *XML Query*.

*Event Log* name filtering is simple:

```toml
  eventlog_name = "Application"
  xpath_query = '''
```

For *XPath Query* filtering set the `xpath_query` value, and `eventlog_name` will be ignored:

```toml
  eventlog_name = ""
  xpath_query = "Event/System[EventID=999]"
```

XML Query is the most flexible: you can Select or Suppress any values, and give ranges for other values.

XML Query documentation is located here:

<https://docs.microsoft.com/en-us/windows/win32/wes/consuming-events>

### Metrics

- win_eventlog
  - tags
    - source (string)
    - event_id (int)
    - level (int)
    - keywords (string): comma-separated in case of multiple values
    - eventlog_name (string)
    - computer (string): only from Forwarded Events
  - fields
    - version (int)
    - task (int)
    - opcode (int): only if not equal to zero
    - record_id (int)
    - time_created (SystemTime)
    - activity_id (string): only if not empty
    - user_id (string): SID
    - process_id (int)
    - process_name (string)
    - thread_id (int)

The `level` tag can have the following values:

- 1 *critical*
- 2 *error*
- 3 *warning*
- 4 *information*
- 5 *verbose*

Keywords are converted from hex uint64 value by the `_EvtFormatMessage` WINAPI function. There can be more than one value, in that case they will be comma-separated. If keywords can't be converted (bad device driver of forwarded from another computer with unknown Event Channel), hex uint64 is saved as is.

### Additional Fields

*Event Data* values from the message XML are added as additional fields automatically: `Name` attribute is taken as the name, and inner text is the value. Type of Event Data Values is always string.

To protect default fields values, if `Name` attribute is the same as some of default fields, it is given a prefix `data_`. E.g. if there is an Event Data field named `Version`, it will become `data_version`.

Event Data fields without Name attributes are added with sequential numbers: `data_1`, `data_2` and so on.

### Localization

Human readable Event Descriptions are skipped in favour of the Event XML values.

Keywords are saved with the current Windows locale by default. You can override this, for example, to English locale by setting `locale` config parameter to `1033`. Unfortunately, Event Data values are in one locale only, default for the current computer, so setting locale value affects only Keywords. Keywords are used as a tag, so it's still useful.

List of locales:

<https://docs.microsoft.com/en-us/openspecs/office_standards/ms-oe376/6c085406-a698-4e12-9d4d-c3b0ee3dbc4a>

### Example Output

```text
win_eventlog,event_id=19,eventlog_name=System,host=PC,keywords=Installation\,Success,level=4,source=Microsoft-Windows-WindowsUpdateClient task=1i,process_id=22284i,thread_id=15220i,user_id="S-1-5-18",updateGuid="{7fc60252-919e-406c-8016-c8202c68dcb3}",serviceGuid="{7971f918-a847-4430-9279-4a52d1efe18d}",version=1i,record_id=1913921i,process_name="svchost.exe",opcode=13i,updateTitle="Обновление механизма обнаружения угроз для Microsoft Defender Antivirus - KB2267602 (версия 1.321.1681.0)",updateRevisionNumber="200" 1597757870000000000

win_eventlog,event_id=4624,eventlog_name=Security,host=PC,keywords=Audit\ Success,level=0,source=Microsoft-Windows-Security-Auditing LogonType="5",ProcessId="0x3bc",ElevatedToken="%%1842",TargetUserName="СИСТЕМА",TargetLogonId="0x3e7",TransmittedServices="-",TargetUserSid="S-1-5-18",WorkstationName="-",ProcessName="C:\\Windows\\System32\\services.exe",VirtualAccount="%%1843",LogonProcessName="Advapi  ",AuthenticationPackageName="Negotiate",IpAddress="-",TargetLinkedLogonId="0x0",SubjectUserSid="S-1-5-18",TargetDomainName="NT AUTHORITY",IpPort="-",RestrictedAdminMode="-",process_name="lsass.exe",SubjectDomainName="WORKGROUP",thread_id=22928i,activity_id="{0d4cc11d-7099-0002-4dc1-4c0d9970d601}",SubjectUserName="PC$",SubjectLogonId="0x3e7",LmPackageName="-",version=2i,task=12544i,TargetOutboundDomainName="-",TargetOutboundUserName="-",record_id=206237i,LogonGuid="{00000000-0000-0000-0000-000000000000}",ImpersonationLevel="%%1833",process_id=996i,KeyLength="0" 1597757860000000000

win_eventlog,event_id=105,eventlog_name=System,host=PC,keywords=0x8000000000000000,level=4,source=WudfUsbccidDriver record_id=1913918i,process_id=17652i,thread_id=21340i,dwMaxCCIDMessageLength="271",version=0i,bClassGetEnvelope="0x0",wLcdLayout="0x0",bPINSupport="0x0",bMaxCCIDBusySlots="1",task=1i,user_id="S-1-5-19",bClassGetResponse="0x0",process_name="WUDFHost.exe",opcode=10i 1597757840000000000
```
