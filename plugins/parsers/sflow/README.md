# SFlow Parser
## Current Scope
V5 not V4
Only Flow samples, no counters
## Future Possibilities
Counters
Other Flow Types
Other Header Types
# Schema
## Tags (optionally fields)
## Fields
# Implementation Approach
Generic packet processing engine - easy to alter, not the most efficient in memory or cpu utilisation due to heavy use of map[string]interface for recording generic object tree.
# Tests
Hard to come by good packets and deconstructed contents
Hesvy use of wireshark and intersection of various open source tools plus heavy readying of spec.

Stoachastic tests to help find edge cases.


USE Telegraf logging - what shoudl I really do re warnings?


