# systemd_timings Input Plugin

The systemd_timings plugin collects timestamps relating to the systemd based
boot process.  All values are accessed via systemd APIs which are exposed
on D-Bus. For more information on the systemd D-Bus API see:

   * https://www.freedesktop.org/wiki/Software/systemd/dbus/

## System Wide Boot Timestamps

The values produced here indicate timestamps of various system wide boot tasks:

   * FirmwareTimestampMonotonic
   * LoaderTimestampMonotonic
   * InitRDTimestampMonotonic
   * UserspaceTimestampMonotonic
   * FinishTimestampMonotonic
   * SecurityStartTimestampMonotonic
   * SecurityFinishTimestampMonotonic
   * GeneratorsStartTimestampMonotonic
   * GeneratorsFinishTimestampMonotonic
   * UnitsLoadStartTimestampMonotonic
   * UnitsLoadFinishTimestampMonotonic
   * InitRDSecurityStartTimestampMonotonic
   * InitRDSecurityFinishTimestampMonotonic
   * InitRDGeneratorsStartTimestampMonotonic
   * InitRDGeneratorsFinishTimestampMonotonic
   * InitRDUnitsLoadStartTimestampMonotonic
   * InitRDUnitsLoadFinishTimestampMonotonic

All values are of type uint64 and are measured in microseconds.  These
timestamps are sent any time telegraf is started.

## Unit Activation/Deactivation Timestamps

For each unit in the system the following timestamps are produced:

   * Activating
   * Activated
   * Deactivating
   * Deactivated
   * Time

The "Time" timestamp is the delta between Activated and Activating OR between
Deactivated and Deactivating depending on which set of timestamps is non zero.
This corresponds to the amount of time that it took a unit to start or to stop.

All values are of type uint64 and are measured in microseconds since userspace
start.  These timestamps are sent for all units on telegraf startup, if a unit
is restarted or stopped whilst telegraf is running then the new versions of the
timestamps are sent.
