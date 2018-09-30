:<<"::CMDLITERAL"
@ECHO OFF
GOTO :BATCHSCRIPT
::CMDLITERAL

################ BASH

echo "executed" > "$1"
exit 0

################ end of BASH

rem ############ BATCH
:BATCHSCRIPT
@echo off

echo EXECUTED > %1
exit /b

rem ############ end of BATCH